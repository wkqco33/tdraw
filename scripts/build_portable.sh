#!/usr/bin/env bash
# 컨테이너 안에서 실행되는 낮은 glibc 베이스라인 비디오 빌드 스크립트.
# 정적 FFmpeg + libtcamviewer.a + tdraw Go 바이너리를 한 파일로 묶는다.
#
# 마운트 전제:
#   /w/tdraw       tdraw 소스 (쓰기 가능)
#   /w/tcamviewer  tcamviewer 소스 (읽기 가능)
#   /w/ffmpeg      정적 FFmpeg prefix 캐시 (쓰기 가능, 없으면 빌드)
# 사용: docker/portable.Dockerfile 이미지 안에서 scripts/build_portable.sh
set -euo pipefail

TDRW_DIR="${TDRW_DIR:-/w/tdraw}"
TCAM_DIR="${TCAM_DIR:-/w/tcamviewer}"
FFMPEG_PREFIX="${TCAMVIEWER_FFMPEG_DIR:-/w/ffmpeg}"
CORE_BUILD_DIR="${CORE_BUILD_DIR:-/w/core-static}"
LINK_DIR="${LINK_DIR:-/w/linkportable}"
OUT_DIR="${OUT_DIR:-$TDRW_DIR/dist/portable}"
JOBS="${JOBS:-$(nproc)}"

echo "[1/4] 정적 FFmpeg 준비: $FFMPEG_PREFIX"
bash "$TCAM_DIR/scripts/build_portable_ffmpeg.sh" "$FFMPEG_PREFIX"

echo "[2/4] 정적 FFmpeg로 libtcamviewer 빌드"
cmake -B "$CORE_BUILD_DIR" -S "$TCAM_DIR" \
    -DCMAKE_BUILD_TYPE=Release \
    -DBUILD_TESTING=OFF \
    -DTCAMVIEWER_FFMPEG_STATIC=ON \
    -DTCAMVIEWER_FFMPEG_DIR="$FFMPEG_PREFIX"
cmake --build "$CORE_BUILD_DIR" --target tcamviewer_static -j "$JOBS"

# Go 링크는 cgo 지시자의 -ltcamviewer -lavcodec ... 를 그대로 쓴다. 순서를 바꿀 수
# 없어 정적 아카이브끼리 의존성이 깨지므로, 코어와 FFmpeg 아카이브를 하나로 합쳐
# libtcamviewer.a로 만들고 나머지 이름은 빈 아카이브로 채운다. 그러면 ld가 .so 대신
# 항상 정적 아카이브를 선택하고 런타임 FFmpeg 의존도 사라진다.
echo "[3/4] 정적 링크 디렉터리 구성: $LINK_DIR"
mkdir -p "$LINK_DIR" "$OUT_DIR"
ar -M <<EOF
create $LINK_DIR/libtcamviewer.a
addlib $CORE_BUILD_DIR/libtcamviewer.a
addlib $FFMPEG_PREFIX/lib/libavformat.a
addlib $FFMPEG_PREFIX/lib/libavcodec.a
addlib $FFMPEG_PREFIX/lib/libswscale.a
addlib $FFMPEG_PREFIX/lib/libswresample.a
addlib $FFMPEG_PREFIX/lib/libavutil.a
save
end
EOF
for lib in avformat avcodec swscale swresample avutil; do
    rm -f "$LINK_DIR/lib${lib}.a"
    ar crs "$LINK_DIR/lib${lib}.a"
done

echo "[4/4] tdraw 빌드"
cd "$TDRW_DIR"
version="${VERSION:-dev}"
GOWORK=off CGO_ENABLED=1 CGO_LDFLAGS="-L$LINK_DIR" \
    go build -tags tcamviewer -trimpath -buildvcs=false \
    -ldflags "-s -w -X main.version=$version" \
    -o "$OUT_DIR/tdraw" ./cmd/tdraw

echo "완료: $OUT_DIR/tdraw"

# 컨테이너 루트로 빌드되었으므로 호스트 사용자 소유권으로 되돌려 준다.
if [ -n "${HOST_UID:-}" ] && [ "$HOST_UID" != "$(id -u)" ]; then
    chown -R "$HOST_UID:${HOST_GID:-$HOST_UID}" "$OUT_DIR"
fi

ldd "$OUT_DIR/tdraw" || true
