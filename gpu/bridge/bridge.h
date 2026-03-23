#ifndef WYPAS_GPU_BRIDGE_H
#define WYPAS_GPU_BRIDGE_H

typedef struct {
    void* device;
    void* queue;
    void* library;
    void* compositePS;
    void* outfitTintPS;
    void* resizeNNPS;
    void* resizeCatmullPS;
    void* heatmapPS;
    void* shadowPS;
    void* minimapPS;
} MetalGPU;

MetalGPU* MetalGPU_Init(const char* shaderSource, int shaderLen);
void MetalGPU_Destroy(MetalGPU* gpu);

int MetalGPU_Composite(MetalGPU* gpu, unsigned char* dst, int dstW, int dstH,
    unsigned char* srcData, int* descriptors, int numOps);

int MetalGPU_TintOutfit(MetalGPU* gpu, unsigned char* base, unsigned char* overlay,
    int w, int h, unsigned char head[3], unsigned char body[3],
    unsigned char legs[3], unsigned char feet[3], int isAddon, unsigned char* out);

int MetalGPU_ResizeNN(MetalGPU* gpu, unsigned char* src, int srcW, int srcH,
    int dstW, int dstH, unsigned char* out);

int MetalGPU_ResizeCatmullRom(MetalGPU* gpu, unsigned char* src, int srcW, int srcH,
    int dstW, int dstH, unsigned char* out);

int MetalGPU_RenderHeatmap(MetalGPU* gpu, float* freq, int imgW, int imgH,
    int radius, unsigned char* out);

int MetalGPU_ApplyShadow(MetalGPU* gpu, unsigned char* pixels, int w, int h, float intensity);

int MetalGPU_RenderMinimap(MetalGPU* gpu, int w, int h,
    int* tileData, int numTiles, unsigned char* out);

#endif
