#import <Metal/Metal.h>
#import <Foundation/Foundation.h>
#include "bridge/bridge.h"
#include <string.h>

MetalGPU* MetalGPU_Init(const char* shaderSource, int shaderLen) {
    @autoreleasepool {
        id<MTLDevice> device = MTLCreateSystemDefaultDevice();
        if (!device) return NULL;

        id<MTLCommandQueue> queue = [device newCommandQueue];
        if (!queue) return NULL;

        NSString* src = [[NSString alloc] initWithBytes:shaderSource
                                                 length:shaderLen
                                               encoding:NSUTF8StringEncoding];
        NSError* error = nil;
        id<MTLLibrary> library = [device newLibraryWithSource:src options:nil error:&error];
        if (!library) {
            NSLog(@"Metal shader compile error: %@", error);
            return NULL;
        }

        NSArray* names = @[
            @"composite_batch", @"outfit_tint", @"resize_nn",
            @"resize_catmull", @"heatmap", @"shadow_overlay", @"minimap_color"
        ];

        id<MTLComputePipelineState> pipelines[7];
        for (int i = 0; i < 7; i++) {
            id<MTLFunction> fn = [library newFunctionWithName:names[i]];
            if (!fn) {
                NSLog(@"Metal function not found: %@", names[i]);
                return NULL;
            }
            pipelines[i] = [device newComputePipelineStateWithFunction:fn error:&error];
            if (!pipelines[i]) {
                NSLog(@"Metal pipeline error: %@", error);
                return NULL;
            }
        }

        MetalGPU* gpu = (MetalGPU*)malloc(sizeof(MetalGPU));
        gpu->device = (__bridge_retained void*)device;
        gpu->queue = (__bridge_retained void*)queue;
        gpu->library = (__bridge_retained void*)library;
        gpu->compositePS = (__bridge_retained void*)pipelines[0];
        gpu->outfitTintPS = (__bridge_retained void*)pipelines[1];
        gpu->resizeNNPS = (__bridge_retained void*)pipelines[2];
        gpu->resizeCatmullPS = (__bridge_retained void*)pipelines[3];
        gpu->heatmapPS = (__bridge_retained void*)pipelines[4];
        gpu->shadowPS = (__bridge_retained void*)pipelines[5];
        gpu->minimapPS = (__bridge_retained void*)pipelines[6];

        return gpu;
    }
}

void MetalGPU_Destroy(MetalGPU* gpu) {
    if (!gpu) return;
    CFRelease(gpu->device);
    CFRelease(gpu->queue);
    CFRelease(gpu->library);
    CFRelease(gpu->compositePS);
    CFRelease(gpu->outfitTintPS);
    CFRelease(gpu->resizeNNPS);
    CFRelease(gpu->resizeCatmullPS);
    CFRelease(gpu->heatmapPS);
    CFRelease(gpu->shadowPS);
    CFRelease(gpu->minimapPS);
    free(gpu);
}

static id<MTLCommandQueue> getQueue(MetalGPU* gpu) {
    return (__bridge id<MTLCommandQueue>)gpu->queue;
}

static id<MTLDevice> getDevice(MetalGPU* gpu) {
    return (__bridge id<MTLDevice>)gpu->device;
}

int MetalGPU_Composite(MetalGPU* gpu, unsigned char* dst, int dstW, int dstH,
    unsigned char* srcData, int* descriptors, int numOps) {
    @autoreleasepool {
        id<MTLDevice> dev = getDevice(gpu);
        int dstSize = dstW * dstH * 4;

        id<MTLBuffer> dstBuf = [dev newBufferWithBytes:dst length:dstSize options:MTLResourceStorageModeShared];
        int srcSize = 0;
        for (int i = 0; i < numOps; i++) {
            int off = descriptors[i*5];
            int w = descriptors[i*5+1];
            int h = descriptors[i*5+2];
            int end = off + w * h * 4;
            if (end > srcSize) srcSize = end;
        }
        id<MTLBuffer> srcBuf = [dev newBufferWithBytes:srcData length:srcSize options:MTLResourceStorageModeShared];
        id<MTLBuffer> descBuf = [dev newBufferWithBytes:descriptors length:numOps*5*sizeof(int) options:MTLResourceStorageModeShared];

        id<MTLCommandBuffer> cb = [getQueue(gpu) commandBuffer];
        id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
        [enc setComputePipelineState:(__bridge id<MTLComputePipelineState>)gpu->compositePS];
        [enc setBuffer:dstBuf offset:0 atIndex:0];
        [enc setBuffer:srcBuf offset:0 atIndex:1];
        [enc setBuffer:descBuf offset:0 atIndex:2];
        [enc setBytes:&dstW length:sizeof(int) atIndex:3];
        [enc setBytes:&dstH length:sizeof(int) atIndex:4];
        [enc setBytes:&numOps length:sizeof(int) atIndex:5];

        MTLSize grid = MTLSizeMake(dstW, dstH, 1);
        MTLSize group = MTLSizeMake(16, 16, 1);
        [enc dispatchThreads:grid threadsPerThreadgroup:group];
        [enc endEncoding];
        [cb commit];
        [cb waitUntilCompleted];

        memcpy(dst, dstBuf.contents, dstSize);
        return 0;
    }
}

int MetalGPU_TintOutfit(MetalGPU* gpu, unsigned char* base, unsigned char* overlay,
    int w, int h, unsigned char head[3], unsigned char body[3],
    unsigned char legs[3], unsigned char feet[3], int isAddon, unsigned char* out) {
    @autoreleasepool {
        id<MTLDevice> dev = getDevice(gpu);
        int size = w * h * 4;

        id<MTLBuffer> baseBuf = [dev newBufferWithBytes:base length:size options:MTLResourceStorageModeShared];
        int hasOverlay = overlay ? 1 : 0;
        id<MTLBuffer> overlayBuf;
        if (overlay) {
            overlayBuf = [dev newBufferWithBytes:overlay length:size options:MTLResourceStorageModeShared];
        } else {
            overlayBuf = [dev newBufferWithLength:4 options:MTLResourceStorageModeShared];
        }
        id<MTLBuffer> outBuf = [dev newBufferWithLength:size options:MTLResourceStorageModeShared];

        unsigned char headC[4] = {head[0], head[1], head[2], 0};
        unsigned char bodyC[4] = {body[0], body[1], body[2], 0};
        unsigned char legsC[4] = {legs[0], legs[1], legs[2], 0};
        unsigned char feetC[4] = {feet[0], feet[1], feet[2], 0};

        id<MTLCommandBuffer> cb = [getQueue(gpu) commandBuffer];
        id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
        [enc setComputePipelineState:(__bridge id<MTLComputePipelineState>)gpu->outfitTintPS];
        [enc setBuffer:baseBuf offset:0 atIndex:0];
        [enc setBuffer:overlayBuf offset:0 atIndex:1];
        [enc setBuffer:outBuf offset:0 atIndex:2];
        [enc setBytes:&w length:sizeof(int) atIndex:3];
        [enc setBytes:&h length:sizeof(int) atIndex:4];
        [enc setBytes:headC length:4 atIndex:5];
        [enc setBytes:bodyC length:4 atIndex:6];
        [enc setBytes:legsC length:4 atIndex:7];
        [enc setBytes:feetC length:4 atIndex:8];
        [enc setBytes:&isAddon length:sizeof(int) atIndex:9];
        [enc setBytes:&hasOverlay length:sizeof(int) atIndex:10];

        MTLSize grid = MTLSizeMake(w, h, 1);
        MTLSize group = MTLSizeMake(16, 16, 1);
        [enc dispatchThreads:grid threadsPerThreadgroup:group];
        [enc endEncoding];
        [cb commit];
        [cb waitUntilCompleted];

        memcpy(out, outBuf.contents, size);
        return 0;
    }
}

int MetalGPU_ResizeNN(MetalGPU* gpu, unsigned char* src, int srcW, int srcH,
    int dstW, int dstH, unsigned char* out) {
    @autoreleasepool {
        id<MTLDevice> dev = getDevice(gpu);
        id<MTLBuffer> srcBuf = [dev newBufferWithBytes:src length:srcW*srcH*4 options:MTLResourceStorageModeShared];
        id<MTLBuffer> dstBuf = [dev newBufferWithLength:dstW*dstH*4 options:MTLResourceStorageModeShared];

        id<MTLCommandBuffer> cb = [getQueue(gpu) commandBuffer];
        id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
        [enc setComputePipelineState:(__bridge id<MTLComputePipelineState>)gpu->resizeNNPS];
        [enc setBuffer:srcBuf offset:0 atIndex:0];
        [enc setBuffer:dstBuf offset:0 atIndex:1];
        [enc setBytes:&srcW length:sizeof(int) atIndex:2];
        [enc setBytes:&srcH length:sizeof(int) atIndex:3];
        [enc setBytes:&dstW length:sizeof(int) atIndex:4];
        [enc setBytes:&dstH length:sizeof(int) atIndex:5];

        MTLSize grid = MTLSizeMake(dstW, dstH, 1);
        MTLSize group = MTLSizeMake(16, 16, 1);
        [enc dispatchThreads:grid threadsPerThreadgroup:group];
        [enc endEncoding];
        [cb commit];
        [cb waitUntilCompleted];

        memcpy(out, dstBuf.contents, dstW*dstH*4);
        return 0;
    }
}

int MetalGPU_ResizeCatmullRom(MetalGPU* gpu, unsigned char* src, int srcW, int srcH,
    int dstW, int dstH, unsigned char* out) {
    @autoreleasepool {
        id<MTLDevice> dev = getDevice(gpu);
        id<MTLBuffer> srcBuf = [dev newBufferWithBytes:src length:srcW*srcH*4 options:MTLResourceStorageModeShared];
        id<MTLBuffer> dstBuf = [dev newBufferWithLength:dstW*dstH*4 options:MTLResourceStorageModeShared];

        id<MTLCommandBuffer> cb = [getQueue(gpu) commandBuffer];
        id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
        [enc setComputePipelineState:(__bridge id<MTLComputePipelineState>)gpu->resizeCatmullPS];
        [enc setBuffer:srcBuf offset:0 atIndex:0];
        [enc setBuffer:dstBuf offset:0 atIndex:1];
        [enc setBytes:&srcW length:sizeof(int) atIndex:2];
        [enc setBytes:&srcH length:sizeof(int) atIndex:3];
        [enc setBytes:&dstW length:sizeof(int) atIndex:4];
        [enc setBytes:&dstH length:sizeof(int) atIndex:5];

        MTLSize grid = MTLSizeMake(dstW, dstH, 1);
        MTLSize group = MTLSizeMake(16, 16, 1);
        [enc dispatchThreads:grid threadsPerThreadgroup:group];
        [enc endEncoding];
        [cb commit];
        [cb waitUntilCompleted];

        memcpy(out, dstBuf.contents, dstW*dstH*4);
        return 0;
    }
}

int MetalGPU_RenderHeatmap(MetalGPU* gpu, float* freq, int imgW, int imgH,
    int radius, unsigned char* out) {
    @autoreleasepool {
        id<MTLDevice> dev = getDevice(gpu);
        int pixelCount = imgW * imgH;

        // Pre-compute maxVal on CPU
        float maxVal = 0;
        float invR2 = 1.0f / (float)(radius * radius);
        for (int py = 0; py < imgH; py++) {
            for (int px = 0; px < imgW; px++) {
                float heat = 0;
                int y0 = py - radius; if (y0 < 0) y0 = 0;
                int y1 = py + radius; if (y1 >= imgH) y1 = imgH - 1;
                int x0 = px - radius; if (x0 < 0) x0 = 0;
                int x1 = px + radius; if (x1 >= imgW) x1 = imgW - 1;
                for (int j = y0; j <= y1; j++) {
                    float dy = (float)(j - py);
                    for (int i = x0; i <= x1; i++) {
                        float dx = (float)(i - px);
                        float d2 = dx*dx + dy*dy;
                        if (d2 > (float)(radius*radius)) continue;
                        float w = expf(-3.0f * d2 * invR2);
                        heat += freq[j * imgW + i] * w;
                    }
                }
                if (heat > maxVal) maxVal = heat;
            }
        }
        if (maxVal <= 0) {
            memset(out, 0, pixelCount * 4);
            return 0;
        }

        id<MTLBuffer> freqBuf = [dev newBufferWithBytes:freq length:pixelCount*sizeof(float) options:MTLResourceStorageModeShared];
        id<MTLBuffer> outBuf = [dev newBufferWithLength:pixelCount*4 options:MTLResourceStorageModeShared];

        id<MTLCommandBuffer> cb = [getQueue(gpu) commandBuffer];
        id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
        [enc setComputePipelineState:(__bridge id<MTLComputePipelineState>)gpu->heatmapPS];
        [enc setBuffer:freqBuf offset:0 atIndex:0];
        [enc setBuffer:outBuf offset:0 atIndex:1];
        [enc setBytes:&imgW length:sizeof(int) atIndex:2];
        [enc setBytes:&imgH length:sizeof(int) atIndex:3];
        [enc setBytes:&radius length:sizeof(int) atIndex:4];
        [enc setBytes:&maxVal length:sizeof(float) atIndex:5];

        MTLSize grid = MTLSizeMake(imgW, imgH, 1);
        MTLSize group = MTLSizeMake(16, 16, 1);
        [enc dispatchThreads:grid threadsPerThreadgroup:group];
        [enc endEncoding];
        [cb commit];
        [cb waitUntilCompleted];

        memcpy(out, outBuf.contents, pixelCount * 4);
        return 0;
    }
}

int MetalGPU_ApplyShadow(MetalGPU* gpu, unsigned char* pixels, int w, int h, float intensity) {
    @autoreleasepool {
        id<MTLDevice> dev = getDevice(gpu);
        int size = w * h * 4;

        id<MTLBuffer> buf = [dev newBufferWithBytes:pixels length:size options:MTLResourceStorageModeShared];

        id<MTLCommandBuffer> cb = [getQueue(gpu) commandBuffer];
        id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
        [enc setComputePipelineState:(__bridge id<MTLComputePipelineState>)gpu->shadowPS];
        [enc setBuffer:buf offset:0 atIndex:0];
        [enc setBytes:&w length:sizeof(int) atIndex:1];
        [enc setBytes:&h length:sizeof(int) atIndex:2];
        [enc setBytes:&intensity length:sizeof(float) atIndex:3];

        MTLSize grid = MTLSizeMake(w, h, 1);
        MTLSize group = MTLSizeMake(16, 16, 1);
        [enc dispatchThreads:grid threadsPerThreadgroup:group];
        [enc endEncoding];
        [cb commit];
        [cb waitUntilCompleted];

        memcpy(pixels, buf.contents, size);
        return 0;
    }
}

int MetalGPU_RenderMinimap(MetalGPU* gpu, int w, int h,
    int* tileData, int numTiles, unsigned char* out) {
    @autoreleasepool {
        id<MTLDevice> dev = getDevice(gpu);
        int size = w * h * 4;

        id<MTLBuffer> tileBuf = [dev newBufferWithBytes:tileData length:numTiles*3*sizeof(int) options:MTLResourceStorageModeShared];
        id<MTLBuffer> outBuf = [dev newBufferWithBytes:out length:size options:MTLResourceStorageModeShared];

        id<MTLCommandBuffer> cb = [getQueue(gpu) commandBuffer];
        id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
        [enc setComputePipelineState:(__bridge id<MTLComputePipelineState>)gpu->minimapPS];
        [enc setBuffer:tileBuf offset:0 atIndex:0];
        [enc setBuffer:outBuf offset:0 atIndex:1];
        [enc setBytes:&w length:sizeof(int) atIndex:2];
        [enc setBytes:&numTiles length:sizeof(int) atIndex:3];

        MTLSize grid = MTLSizeMake(numTiles, 1, 1);
        MTLSize group = MTLSizeMake(256, 1, 1);
        [enc dispatchThreads:grid threadsPerThreadgroup:group];
        [enc endEncoding];
        [cb commit];
        [cb waitUntilCompleted];

        memcpy(out, outBuf.contents, size);
        return 0;
    }
}
