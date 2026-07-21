#!/usr/bin/env bash

set -Eeuo pipefail

appName="alist"
mode="${1:-}"
target="${2:-}"
ldflags=""

case "$mode" in
dev)
  version="dev"
  webVersion="dev"
  ;;
beta)
  version="beta"
  webVersion="dev"
  ;;
release)
  git tag -d beta >/dev/null 2>&1 || true
  version=$(git describe --abbrev=0 --tags)
  webReleaseUrl=$(curl --fail --silent --show-error --location --head --retry 3 \
    --output /dev/null --write-out '%{url_effective}' \
    "https://github.com/AlliotTech/openalist-web/releases/latest")
  webVersion="${webReleaseUrl##*/}"
  if [[ -z "$webVersion" ]]; then
    echo "Failed to resolve the latest frontend version" >&2
    exit 1
  fi
  ;;
prepare | zip)
  ;;
*)
  echo "Parameter error" >&2
  exit 2
  ;;
esac

if [[ "$mode" == "dev" || "$mode" == "beta" || "$mode" == "release" ]]; then
  builtAt="$(date +'%F %T %z')"
  gitAuthor="Xhofe <i@nn.ci>"
  gitCommit=$(git log --pretty=format:"%h" -1)

  echo "backend version: $version"
  echo "frontend version: $webVersion"

  ldflags="\
-w -s \
-X 'github.com/AlliotTech/openalist/internal/conf.BuiltAt=$builtAt' \
-X 'github.com/AlliotTech/openalist/internal/conf.GitAuthor=$gitAuthor' \
-X 'github.com/AlliotTech/openalist/internal/conf.GitCommit=$gitCommit' \
-X 'github.com/AlliotTech/openalist/internal/conf.Version=$version' \
-X 'github.com/AlliotTech/openalist/internal/conf.WebVersion=$webVersion' \
"
fi

FetchWebDev() {
  curl --fail --location --retry 3 \
    https://codeload.github.com/AlliotTech/web-dist/tar.gz/refs/heads/dev \
    --output web-dist-dev.tar.gz
  tar -zxvf web-dist-dev.tar.gz
  rm -rf public/dist
  mv -f web-dist-dev/dist public
  rm -rf web-dist-dev web-dist-dev.tar.gz
}

FetchWebRelease() {
  curl --fail --location --retry 3 \
    https://github.com/AlliotTech/openalist-web/releases/latest/download/dist.tar.gz \
    --output dist.tar.gz
  tar -zxvf dist.tar.gz
  rm -rf public/dist
  mv -f dist public
  rm -rf dist.tar.gz
}

BuildWinArm64() {
  echo building for windows-arm64
  chmod +x ./wrapper/zcc-arm64
  chmod +x ./wrapper/zcxx-arm64
  export GOOS=windows
  export GOARCH=arm64
  export CC="$PWD/wrapper/zcc-arm64"
  export CXX="$PWD/wrapper/zcxx-arm64"
  export CGO_ENABLED=1
  go build -o "$1" -ldflags="$ldflags" -tags=jsoniter .
}

BuildDev() {
  rm -rf .git/
  mkdir -p "dist"
  muslflags="--extldflags '-static -fpic' $ldflags"
  BASE="https://musl.nn.ci/"
  FILES=(x86_64-linux-musl-cross aarch64-linux-musl-cross)
  for i in "${FILES[@]}"; do
    url="${BASE}${i}.tgz"
    curl --fail --location --retry 3 --output "${i}.tgz" "${url}"
    sudo tar xf "${i}.tgz" --strip-components 1 -C /usr/local
  done
  OS_ARCHES=(linux-musl-amd64 linux-musl-arm64)
  CGO_ARGS=(x86_64-linux-musl-gcc aarch64-linux-musl-gcc)
  for i in "${!OS_ARCHES[@]}"; do
    os_arch=${OS_ARCHES[$i]}
    cgo_cc=${CGO_ARGS[$i]}
    echo "building for ${os_arch}"
    export GOOS=${os_arch%%-*}
    export GOARCH=${os_arch##*-}
    export CC=${cgo_cc}
    export CGO_ENABLED=1
    go build -o "./dist/${appName}-${os_arch}" -ldflags="$muslflags" -tags=jsoniter .
  done
  xgo -targets=windows/amd64,darwin/amd64,darwin/arm64 -out "$appName" -ldflags="$ldflags" -tags=jsoniter .
  mv alist-* dist
  cd dist
  cp ./alist-windows-amd64.exe ./alist-windows-amd64-upx.exe
  upx -9 ./alist-windows-amd64-upx.exe
  find . -type f -print0 | xargs -0 md5sum >md5.txt
  cat md5.txt
}

BuildDocker() {
  go build -o ./bin/alist -ldflags="$ldflags" -tags=jsoniter .
}

PrepareBuildDockerMusl() {
  mkdir -p build/musl-libs
  # BASE="https://musl.cc/"
  BASE="https://musl.nn.ci/"
  FILES=(x86_64-linux-musl-cross aarch64-linux-musl-cross)
  # FILES=(x86_64-linux-musl-cross aarch64-linux-musl-cross armv6-linux-musleabihf-cross armv7l-linux-musleabihf-cross)
  for i in "${FILES[@]}"; do
    url="${BASE}${i}.tgz"
    lib_tgz="build/${i}.tgz"
    curl --fail --location --retry 3 --output "${lib_tgz}" "${url}"
    tar xf "${lib_tgz}" --strip-components 1 -C build/musl-libs
    rm -f "${lib_tgz}"
  done
}

BuildDockerMultiplatform() {
  go mod download

  # run PrepareBuildDockerMusl before build
  export PATH=$PATH:$PWD/build/musl-libs/bin

  docker_lflags="--extldflags '-static -fpic' $ldflags"
  export CGO_ENABLED=1

  OS_ARCHES=(linux-amd64 linux-arm64)
  CGO_ARGS=(x86_64-linux-musl-gcc aarch64-linux-musl-gcc i486-linux-musl-gcc s390x-linux-musl-gcc riscv64-linux-musl-gcc)
  for i in "${!OS_ARCHES[@]}"; do
    os_arch=${OS_ARCHES[$i]}
    cgo_cc=${CGO_ARGS[$i]}
    os=${os_arch%%-*}
    arch=${os_arch##*-}
    export GOOS=$os
    export GOARCH=$arch
    export CC=${cgo_cc}
    echo "building for $os_arch"
    go build -o "build/$os/$arch/alist" -ldflags="$docker_lflags" -tags=jsoniter .
  done

  # DOCKER_ARM_ARCHES=(linux-arm/v6 linux-arm/v7)
  # CGO_ARGS=(armv6-linux-musleabihf-gcc armv7l-linux-musleabihf-gcc)
  # GO_ARM=(6 7)
  # export GOOS=linux
  # export GOARCH=arm
  # for i in "${!DOCKER_ARM_ARCHES[@]}"; do
  #   docker_arch=${DOCKER_ARM_ARCHES[$i]}
  #   cgo_cc=${CGO_ARGS[$i]}
  #   export GOARM=${GO_ARM[$i]}
  #   export CC=${cgo_cc}
  #   echo "building for $docker_arch"
  #   go build -o build/${docker_arch%%-*}/${docker_arch##*-}/alist -ldflags="$docker_lflags" -tags=jsoniter .
  # done
}

BuildRelease() {
  rm -rf .git/
  mkdir -p "build"
  BuildWinArm64 ./build/alist-windows-arm64.exe
  xgo -out "$appName" -ldflags="$ldflags" -tags=jsoniter .
  # why? Because some target platforms seem to have issues with upx compression
  upx -9 ./alist-linux-amd64
  cp ./alist-windows-amd64.exe ./alist-windows-amd64-upx.exe
  upx -9 ./alist-windows-amd64-upx.exe
  mv alist-* build
}

BuildReleaseLinuxMusl() {
  rm -rf .git/
  mkdir -p "build"
  muslflags="--extldflags '-static -fpic' $ldflags"
  BASE="https://musl.nn.ci/"
  FILES=(x86_64-linux-musl-cross aarch64-linux-musl-cross mips-linux-musl-cross mips64-linux-musl-cross mips64el-linux-musl-cross mipsel-linux-musl-cross powerpc64le-linux-musl-cross s390x-linux-musl-cross)
  for i in "${FILES[@]}"; do
    url="${BASE}${i}.tgz"
    curl --fail --location --retry 3 --output "${i}.tgz" "${url}"
    sudo tar xf "${i}.tgz" --strip-components 1 -C /usr/local
    rm -f "${i}.tgz"
  done
  OS_ARCHES=(linux-musl-amd64 linux-musl-arm64 linux-musl-mips linux-musl-mips64 linux-musl-mips64le linux-musl-mipsle linux-musl-ppc64le linux-musl-s390x)
  CGO_ARGS=(x86_64-linux-musl-gcc aarch64-linux-musl-gcc mips-linux-musl-gcc mips64-linux-musl-gcc mips64el-linux-musl-gcc mipsel-linux-musl-gcc powerpc64le-linux-musl-gcc s390x-linux-musl-gcc)
  for i in "${!OS_ARCHES[@]}"; do
    os_arch=${OS_ARCHES[$i]}
    cgo_cc=${CGO_ARGS[$i]}
    echo "building for ${os_arch}"
    export GOOS=${os_arch%%-*}
    export GOARCH=${os_arch##*-}
    export CC=${cgo_cc}
    export CGO_ENABLED=1
    go build -o "./build/${appName}-${os_arch}" -ldflags="$muslflags" -tags=jsoniter .
  done
}

BuildReleaseLinuxMuslArm() {
  rm -rf .git/
  mkdir -p "build"
  muslflags="--extldflags '-static -fpic' $ldflags"
  BASE="https://musl.nn.ci/"
#  FILES=(arm-linux-musleabi-cross arm-linux-musleabihf-cross armeb-linux-musleabi-cross armeb-linux-musleabihf-cross armel-linux-musleabi-cross armel-linux-musleabihf-cross armv5l-linux-musleabi-cross armv5l-linux-musleabihf-cross armv6-linux-musleabi-cross armv6-linux-musleabihf-cross armv7l-linux-musleabihf-cross armv7m-linux-musleabi-cross armv7r-linux-musleabihf-cross)
  FILES=(arm-linux-musleabi-cross arm-linux-musleabihf-cross armel-linux-musleabi-cross armel-linux-musleabihf-cross armv5l-linux-musleabi-cross armv5l-linux-musleabihf-cross armv6-linux-musleabi-cross armv6-linux-musleabihf-cross armv7l-linux-musleabihf-cross armv7m-linux-musleabi-cross armv7r-linux-musleabihf-cross)
  for i in "${FILES[@]}"; do
    url="${BASE}${i}.tgz"
    curl --fail --location --retry 3 --output "${i}.tgz" "${url}"
    sudo tar xf "${i}.tgz" --strip-components 1 -C /usr/local
    rm -f "${i}.tgz"
  done
#  OS_ARCHES=(linux-musleabi-arm linux-musleabihf-arm linux-musleabi-armeb linux-musleabihf-armeb linux-musleabi-armel linux-musleabihf-armel linux-musleabi-armv5l linux-musleabihf-armv5l linux-musleabi-armv6 linux-musleabihf-armv6 linux-musleabihf-armv7l linux-musleabi-armv7m linux-musleabihf-armv7r)
#  CGO_ARGS=(arm-linux-musleabi-gcc arm-linux-musleabihf-gcc armeb-linux-musleabi-gcc armeb-linux-musleabihf-gcc armel-linux-musleabi-gcc armel-linux-musleabihf-gcc armv5l-linux-musleabi-gcc armv5l-linux-musleabihf-gcc armv6-linux-musleabi-gcc armv6-linux-musleabihf-gcc armv7l-linux-musleabihf-gcc armv7m-linux-musleabi-gcc armv7r-linux-musleabihf-gcc)
#  GOARMS=('' '' '' '' '' '' '5' '5' '6' '6' '7' '7' '7')
  OS_ARCHES=(linux-musleabi-arm linux-musleabihf-arm linux-musleabi-armel linux-musleabihf-armel linux-musleabi-armv5l linux-musleabihf-armv5l linux-musleabi-armv6 linux-musleabihf-armv6 linux-musleabihf-armv7l linux-musleabi-armv7m linux-musleabihf-armv7r)
  CGO_ARGS=(arm-linux-musleabi-gcc arm-linux-musleabihf-gcc armel-linux-musleabi-gcc armel-linux-musleabihf-gcc armv5l-linux-musleabi-gcc armv5l-linux-musleabihf-gcc armv6-linux-musleabi-gcc armv6-linux-musleabihf-gcc armv7l-linux-musleabihf-gcc armv7m-linux-musleabi-gcc armv7r-linux-musleabihf-gcc)
  GOARMS=('' '' '' '' '5' '5' '6' '6' '7' '7' '7')
  for i in "${!OS_ARCHES[@]}"; do
    os_arch=${OS_ARCHES[$i]}
    cgo_cc=${CGO_ARGS[$i]}
    arm=${GOARMS[$i]}
    echo "building for ${os_arch}"
    export GOOS=linux
    export GOARCH=arm
    export CC=${cgo_cc}
    export CGO_ENABLED=1
    export GOARM=${arm}
    go build -o "./build/${appName}-${os_arch}" -ldflags="$muslflags" -tags=jsoniter .
  done
}

BuildReleaseAndroid() {
  rm -rf .git/
  mkdir -p "build"
  wget https://dl.google.com/android/repository/android-ndk-r26b-linux.zip
  unzip android-ndk-r26b-linux.zip
  rm android-ndk-r26b-linux.zip
  OS_ARCHES=(amd64 arm64 386 arm)
  CGO_ARGS=(x86_64-linux-android24-clang aarch64-linux-android24-clang i686-linux-android24-clang armv7a-linux-androideabi24-clang)
  for i in "${!OS_ARCHES[@]}"; do
    os_arch=${OS_ARCHES[$i]}
    cgo_cc=$(realpath "android-ndk-r26b/toolchains/llvm/prebuilt/linux-x86_64/bin/${CGO_ARGS[$i]}")
    echo "building for android-${os_arch}"
    export GOOS=android
    export GOARCH=${os_arch##*-}
    export CC=${cgo_cc}
    export CGO_ENABLED=1
    go build -o "./build/${appName}-android-${os_arch}" -ldflags="$ldflags" -tags=jsoniter .
    android-ndk-r26b/toolchains/llvm/prebuilt/linux-x86_64/bin/llvm-strip "./build/${appName}-android-${os_arch}"
  done
}

BuildReleaseFreeBSD() {
  rm -rf .git/
  mkdir -p "build/freebsd"
  OS_ARCHES=(amd64 arm64 i386)
  GO_ARCHES=(amd64 arm64 386)
  CGO_ARGS=(x86_64-unknown-freebsd14.1 aarch64-unknown-freebsd14.1 i386-unknown-freebsd14.1)
  for i in "${!OS_ARCHES[@]}"; do
    os_arch=${OS_ARCHES[$i]}
    cgo_cc="clang --target=${CGO_ARGS[$i]} --sysroot=/opt/freebsd/${os_arch}"
    echo "building for freebsd-${os_arch}"
    sudo mkdir -p "/opt/freebsd/${os_arch}"
    wget -q "https://download.freebsd.org/releases/${os_arch}/14.1-RELEASE/base.txz"
    sudo tar -xf ./base.txz -C "/opt/freebsd/${os_arch}"
    rm base.txz
    export GOOS=freebsd
    export GOARCH=${GO_ARCHES[$i]}
    export CC=${cgo_cc}
    export CGO_ENABLED=1
    export CGO_LDFLAGS="-fuse-ld=lld"
    go build -o "./build/${appName}-freebsd-${os_arch}" -ldflags="$ldflags" -tags=jsoniter .
  done
}

MakeRelease() {
  cd build
  mkdir -p compress
  local platform
  local release_file
  local archive_file
  for platform in linux android darwin freebsd; do
    while IFS= read -r -d '' release_file; do
      cp "$release_file" alist
      tar -czvf "compress/${release_file}.tar.gz" alist
      rm -f alist
    done < <(find . -type f -name "${appName}-${platform}-*" -print0)
  done
  while IFS= read -r -d '' release_file; do
    cp "$release_file" alist.exe
    archive_file="${release_file%.*}"
    zip "compress/${archive_file}.zip" alist.exe
    rm -f alist.exe
  done < <(find . -type f -name "${appName}-windows-*" -print0)
  cd compress
  find . -type f -print0 | xargs -0 md5sum >"$1"
  cat "$1"
  cd ../..
}

if [[ "$mode" == "dev" ]]; then
  FetchWebDev
  if [[ "$target" == "docker" ]]; then
    BuildDocker
  elif [[ "$target" == "docker-multiplatform" ]]; then
    BuildDockerMultiplatform
  elif [[ "$target" == "web" ]]; then
    echo "web only"
  else
    BuildDev
  fi
elif [[ "$mode" == "release" || "$mode" == "beta" ]]; then
  if [[ "$mode" == "beta" ]]; then
    FetchWebDev
  else
    FetchWebRelease
  fi
  if [[ "$target" == "docker" ]]; then
    BuildDocker
  elif [[ "$target" == "docker-multiplatform" ]]; then
    BuildDockerMultiplatform
  elif [[ "$target" == "linux_musl_arm" ]]; then
    BuildReleaseLinuxMuslArm
    MakeRelease "md5-linux-musl-arm.txt"
  elif [[ "$target" == "linux_musl" ]]; then
    BuildReleaseLinuxMusl
    MakeRelease "md5-linux-musl.txt"
  elif [[ "$target" == "android" ]]; then
    BuildReleaseAndroid
    MakeRelease "md5-android.txt"
  elif [[ "$target" == "freebsd" ]]; then
    BuildReleaseFreeBSD
    MakeRelease "md5-freebsd.txt"
  elif [[ "$target" == "web" ]]; then
    echo "web only"
  else
    BuildRelease
    MakeRelease "md5.txt"
  fi
elif [[ "$mode" == "prepare" ]]; then
  if [[ "$target" == "docker-multiplatform" ]]; then
    PrepareBuildDockerMusl
  else
    echo "Parameter error" >&2
    exit 2
  fi
elif [[ "$mode" == "zip" && -n "$target" ]]; then
  MakeRelease "${target}.txt"
else
  echo "Parameter error" >&2
  exit 2
fi
