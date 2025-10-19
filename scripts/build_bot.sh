#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<EOF
Usage: $(basename "$0") --context <dir> --image <repository:tag> [--platform <platform>]

Builds a bot container image from the provided context and pushes it to the
supplied Docker repository reference.

Arguments:
  --context   Directory containing the bot Dockerfile (required)
  --image     Fully-qualified image reference to push (e.g., registry/bot:tag)
  --platform  Optional target platform passed to docker buildx (default: local)

Environment:
  BOT_BUILD_ARGS  Additional docker build arguments (space separated)

Examples:
  $(basename "$0") --context bots/python/samples/mean-reversion --image registry.local/bots/mean-reversion:1.0.0
  BOT_BUILD_ARGS="--build-arg VERSION=1.2.3" $(basename "$0") --context . --image registry/bot:latest
EOF
}

context=""
image=""
platform=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --context)
      context="$2"
      shift 2
      ;;
    --image)
      image="$2"
      shift 2
      ;;
    --platform)
      platform="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$context" || -z "$image" ]]; then
  echo "error: --context and --image are required" >&2
  usage
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker CLI not found in PATH" >&2
  exit 1
fi

if [[ ! -d "$context" ]]; then
  echo "error: context directory '$context' does not exist" >&2
  exit 1
fi

echo "Building bot image '$image' from context '$context'"
build_cmd=(docker build "$context" -t "$image")

if [[ -n "$platform" ]]; then
  build_cmd+=(--platform "$platform")
fi

if [[ -n "${BOT_BUILD_ARGS:-}" ]]; then
  read -r -a extra_args <<<"${BOT_BUILD_ARGS}"
  build_cmd+=("${extra_args[@]}")
fi

echo "+ ${build_cmd[*]}"
"${build_cmd[@]}"

echo "Pushing image '$image'"
push_cmd=(docker push "$image")
echo "+ ${push_cmd[*]}"
"${push_cmd[@]}"

echo "Build and push completed successfully"
