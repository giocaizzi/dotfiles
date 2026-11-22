// Jump-only cursor trail (>3 cells, e.g. tmux pane switches) + brief
// inner-edge pulse when a Ghostty native split gains focus.
// Uses Ghostty 1.4 built-in uniforms (prepended by shadertoy_prefix.glsl).

const float MIN_JUMP_CELLS = 3.0;
const float TRAIL_DURATION = 0.22;
const float FOCUS_DURATION = 0.25;

vec2 rectCenter(vec4 r) {
    return vec2(r.x + r.z * 0.5, r.y - r.w * 0.5);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 base = texture(iChannel0, uv);

    if (iFocus == 0) {
        fragColor = base;
        return;
    }

    float trailAlpha = 0.0;
    if (iCursorVisible != 0 && iCurrentCursor.z > 0.0 && iPreviousCursor.z > 0.0) {
        float dt = iTime - iTimeCursorChange;
        if (dt >= 0.0 && dt < TRAIL_DURATION) {
            vec2 curCenter = rectCenter(iCurrentCursor);
            vec2 prevCenter = rectCenter(iPreviousCursor);
            float cellW = max(iCurrentCursor.z, iCurrentCursor.w * 0.5);
            float cellH = max(iCurrentCursor.w, 1.0);
            float cellDist = length((curCenter - prevCenter) / vec2(cellW, cellH));

            if (cellDist > MIN_JUMP_CELLS) {
                float t = clamp(dt / TRAIL_DURATION, 0.0, 1.0);
                float ease = 1.0 - pow(1.0 - t, 3.0);
                vec2 tail = mix(prevCenter, curCenter, ease);
                vec2 ba = curCenter - tail;
                float segLenSq = dot(ba, ba);

                if (segLenSq > 1e-4) {
                    vec2 pa = fragCoord - tail;
                    float h = clamp(dot(pa, ba) / segLenSq, 0.0, 1.0);
                    float dist = length(pa - ba * h);
                    vec2 dir = ba * inversesqrt(segLenSq);
                    vec2 perp = vec2(-dir.y, dir.x);
                    float rHead = dot(abs(perp), vec2(cellW, cellH) * 0.5);
                    float r = mix(rHead * 0.18, rHead, h);

                    float core = 1.0 - smoothstep(r - 1.5, r + 1.5, dist);
                    float glow = exp(-max(dist - r, 0.0) / max(rHead * 1.4, 1.0)) * 0.35;
                    float fade = pow(h, 0.7) * (1.0 - t);

                    // Keep the character under the destination cursor cell sharp.
                    vec2 dCur = abs(fragCoord - curCenter) - iCurrentCursor.zw * 0.5;
                    float curSdf = length(max(dCur, 0.0)) + min(max(dCur.x, dCur.y), 0.0);
                    float outsideCur = smoothstep(-0.5, 1.0, curSdf);

                    trailAlpha = clamp((core * 0.85 + glow) * fade, 0.0, 0.85) * outsideCur;
                }
            }
        }
    }

    float focusAlpha = 0.0;
    float ft = iTime - iTimeFocus;
    if (ft >= 0.0 && ft < FOCUS_DURATION) {
        float fp = ft / FOCUS_DURATION;
        float fFade = (1.0 - fp) * (1.0 - fp);
        float edgeDist = min(
            min(fragCoord.x, iResolution.x - fragCoord.x),
            min(fragCoord.y, iResolution.y - fragCoord.y)
        );
        float borderLine = 1.0 - smoothstep(2.0, 4.0, edgeDist);
        float innerGlow = exp(-edgeDist / 22.0) * 0.22;
        focusAlpha = clamp((borderLine * 0.75 + innerGlow) * fFade, 0.0, 0.75);
    }

    float effectAlpha = clamp(max(trailAlpha, focusAlpha), 0.0, 0.85);
    if (effectAlpha <= 0.001) {
        fragColor = base;
        return;
    }

    vec3 accent = (length(iCurrentCursorColor.rgb) > 0.01)
        ? iCurrentCursorColor.rgb
        : iCursorColor;

    // Porter-Duff source-over preserves background-opacity transparency cleanly.
    fragColor = vec4(
        accent * effectAlpha + base.rgb * (1.0 - effectAlpha),
        effectAlpha + base.a * (1.0 - effectAlpha)
    );
}
