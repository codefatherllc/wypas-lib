#include <metal_stdlib>
using namespace metal;

struct BlitDescriptor {
    int srcOffset;
    int srcW;
    int srcH;
    int dstX;
    int dstY;
};

kernel void composite_batch(
    device uchar4* dst         [[buffer(0)]],
    device const uchar4* src   [[buffer(1)]],
    device const BlitDescriptor* descs [[buffer(2)]],
    constant int& dstW         [[buffer(3)]],
    constant int& dstH         [[buffer(4)]],
    constant int& numOps       [[buffer(5)]],
    uint2 gid                  [[thread_position_in_grid]])
{
    int dx = int(gid.x);
    int dy = int(gid.y);
    if (dx >= dstW || dy >= dstH) return;

    int di = dy * dstW + dx;
    float4 dstC = float4(dst[di]) / 255.0;

    for (int i = 0; i < numOps; i++) {
        BlitDescriptor d = descs[i];
        int sx = dx - d.dstX;
        int sy = dy - d.dstY;
        if (sx < 0 || sx >= d.srcW || sy < 0 || sy >= d.srcH) continue;

        int si = d.srcOffset / 4 + sy * d.srcW + sx;
        float4 srcC = float4(src[si]) / 255.0;
        float sa = srcC.a;
        if (sa <= 0.0) continue;

        float outA = sa + dstC.a * (1.0 - sa);
        if (outA <= 0.0) continue;
        dstC.rgb = (srcC.rgb * sa + dstC.rgb * dstC.a * (1.0 - sa)) / outA;
        dstC.a = outA;
    }

    dst[di] = uchar4(clamp(dstC * 255.0, 0.0, 255.0));
}

kernel void outfit_tint(
    device const uchar4* base     [[buffer(0)]],
    device const uchar4* overlay  [[buffer(1)]],
    device uchar4* out            [[buffer(2)]],
    constant int& w               [[buffer(3)]],
    constant int& h               [[buffer(4)]],
    constant uchar4& headColor    [[buffer(5)]],
    constant uchar4& bodyColor    [[buffer(6)]],
    constant uchar4& legsColor    [[buffer(7)]],
    constant uchar4& feetColor    [[buffer(8)]],
    constant int& isAddon         [[buffer(9)]],
    constant int& hasOverlay      [[buffer(10)]],
    uint2 gid                     [[thread_position_in_grid]])
{
    int x = int(gid.x);
    int y = int(gid.y);
    if (x >= w || y >= h) return;

    int i = y * w + x;
    uchar4 bp = base[i];

    if (bp.a == 0 || (bp.r == 252 && bp.g == 0 && bp.b == 252)) {
        out[i] = uchar4(0);
        return;
    }

    if (hasOverlay == 0) {
        out[i] = bp;
        return;
    }

    uchar4 op = overlay[i];
    uchar3 pal;
    bool matched = true;

    if (op.r == 255 && op.g == 255 && op.b == 0) {
        pal = uchar3(headColor.r, headColor.g, headColor.b);
    } else if (op.r == 255 && op.g == 0 && op.b == 0) {
        pal = uchar3(bodyColor.r, bodyColor.g, bodyColor.b);
    } else if (op.r == 0 && op.g == 255 && op.b == 0) {
        pal = uchar3(legsColor.r, legsColor.g, legsColor.b);
    } else if (op.r == 0 && op.g == 0 && op.b == 255) {
        pal = uchar3(feetColor.r, feetColor.g, feetColor.b);
    } else {
        matched = false;
    }

    if (isAddon != 0) {
        out[i] = uchar4(bp.r, bp.g, bp.b, 255);
        if (matched) {
            out[i] = uchar4(
                uchar(uint(bp.r) * uint(pal.r) / 255),
                uchar(uint(bp.g) * uint(pal.g) / 255),
                uchar(uint(bp.b) * uint(pal.b) / 255),
                255);
        }
    } else {
        if (matched) {
            out[i] = uchar4(
                uchar(uint(bp.r) * uint(pal.r) / 255),
                uchar(uint(bp.g) * uint(pal.g) / 255),
                uchar(uint(bp.b) * uint(pal.b) / 255),
                255);
        } else {
            out[i] = uchar4(bp.r, bp.g, bp.b, 255);
        }
    }
}

kernel void resize_nn(
    device const uchar4* src [[buffer(0)]],
    device uchar4* dst       [[buffer(1)]],
    constant int& srcW       [[buffer(2)]],
    constant int& srcH       [[buffer(3)]],
    constant int& dstW       [[buffer(4)]],
    constant int& dstH       [[buffer(5)]],
    uint2 gid                [[thread_position_in_grid]])
{
    int dx = int(gid.x);
    int dy = int(gid.y);
    if (dx >= dstW || dy >= dstH) return;

    int sx = dx * srcW / dstW;
    int sy = dy * srcH / dstH;
    dst[dy * dstW + dx] = src[sy * srcW + sx];
}

float catmullRomWeight(float t) {
    float at = abs(t);
    if (at <= 1.0) {
        return 1.5 * at * at * at - 2.5 * at * at + 1.0;
    } else if (at <= 2.0) {
        return -0.5 * at * at * at + 2.5 * at * at - 4.0 * at + 2.0;
    }
    return 0.0;
}

kernel void resize_catmull(
    device const uchar4* src [[buffer(0)]],
    device uchar4* dst       [[buffer(1)]],
    constant int& srcW       [[buffer(2)]],
    constant int& srcH       [[buffer(3)]],
    constant int& dstW       [[buffer(4)]],
    constant int& dstH       [[buffer(5)]],
    uint2 gid                [[thread_position_in_grid]])
{
    int dx = int(gid.x);
    int dy = int(gid.y);
    if (dx >= dstW || dy >= dstH) return;

    float sx = (float(dx) + 0.5) * float(srcW) / float(dstW) - 0.5;
    float sy = (float(dy) + 0.5) * float(srcH) / float(dstH) - 0.5;

    int ix = int(floor(sx));
    int iy = int(floor(sy));
    float fx = sx - float(ix);
    float fy = sy - float(iy);

    float4 result = float4(0.0);
    float wSum = 0.0;

    for (int j = -1; j <= 2; j++) {
        for (int i = -1; i <= 2; i++) {
            int cx = clamp(ix + i, 0, srcW - 1);
            int cy = clamp(iy + j, 0, srcH - 1);
            float w = catmullRomWeight(float(i) - fx) * catmullRomWeight(float(j) - fy);
            result += float4(src[cy * srcW + cx]) * w;
            wSum += w;
        }
    }

    if (wSum > 0.0) result /= wSum;
    dst[dy * dstW + dx] = uchar4(clamp(result, 0.0, 255.0));
}

kernel void heatmap(
    device const float* freq  [[buffer(0)]],
    device uchar4* out        [[buffer(1)]],
    constant int& imgW        [[buffer(2)]],
    constant int& imgH        [[buffer(3)]],
    constant int& radius      [[buffer(4)]],
    constant float& maxVal    [[buffer(5)]],
    uint2 gid                 [[thread_position_in_grid]])
{
    int px = int(gid.x);
    int py = int(gid.y);
    if (px >= imgW || py >= imgH) return;

    float r2 = float(radius * radius);
    float invR2 = 1.0 / r2;
    float heat = 0.0;

    int y0 = max(0, py - radius);
    int y1 = min(imgH - 1, py + radius);
    int x0 = max(0, px - radius);
    int x1 = min(imgW - 1, px + radius);

    for (int j = y0; j <= y1; j++) {
        float dy = float(j - py);
        for (int i = x0; i <= x1; i++) {
            float dx = float(i - px);
            float d2 = dx * dx + dy * dy;
            if (d2 > r2) continue;
            float w = exp(-3.0 * d2 * invR2);
            heat += freq[j * imgW + i] * w;
        }
    }

    if (heat <= 0.0 || maxVal <= 0.0) {
        out[py * imgW + px] = uchar4(0);
        return;
    }

    float t = heat / maxVal;
    if (t > 1.0) t = 1.0;

    float3 colors[4] = {
        float3(0, 80, 120),
        float3(0, 255, 255),
        float3(255, 0, 0),
        float3(255, 255, 0)
    };
    float positions[4] = { 0.0, 0.33, 0.66, 1.0 };

    int lo = 0, hi = 1;
    for (int i = 1; i < 4; i++) {
        if (t <= positions[i]) {
            lo = i - 1;
            hi = i;
            break;
        }
    }
    float f = (t - positions[lo]) / (positions[hi] - positions[lo]);
    float3 c = mix(colors[lo], colors[hi], f);
    float alpha = 160.0 + 95.0 * t;

    out[py * imgW + px] = uchar4(uchar(c.r), uchar(c.g), uchar(c.b), uchar(alpha));
}

kernel void shadow_overlay(
    device uchar4* pixels [[buffer(0)]],
    constant int& w       [[buffer(1)]],
    constant int& h       [[buffer(2)]],
    constant float& intensity [[buffer(3)]],
    uint2 gid             [[thread_position_in_grid]])
{
    int x = int(gid.x);
    int y = int(gid.y);
    if (x >= w || y >= h) return;

    int i = y * w + x;
    uchar4 p = pixels[i];
    if (p.a == 0) return;

    float mult = 1.0 - intensity;
    pixels[i] = uchar4(
        uchar(float(p.r) * mult),
        uchar(float(p.g) * mult),
        uchar(float(p.b) * mult),
        p.a);
}

struct MinimapTileDesc {
    int x;
    int y;
    int colorIndex;
};

kernel void minimap_color(
    device const MinimapTileDesc* tiles [[buffer(0)]],
    device uchar4* out                  [[buffer(1)]],
    constant int& w                     [[buffer(2)]],
    constant int& numTiles              [[buffer(3)]],
    uint gid                            [[thread_position_in_grid]])
{
    if (int(gid) >= numTiles) return;

    MinimapTileDesc t = tiles[gid];
    int ci = t.colorIndex;
    uchar r = uchar((ci / 36 % 6) * 51);
    uchar g = uchar((ci / 6 % 6) * 51);
    uchar b = uchar((ci % 6) * 51);

    out[t.y * w + t.x] = uchar4(r, g, b, 255);
}
