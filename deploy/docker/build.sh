#!/bin/bash
# Docker build script for lumi-go (Go Microservice Template)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="${SERVICE_NAME:-lumi-go}"
REGISTRY="${DOCKER_REGISTRY:-lumitut}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
BUILD_TIME="${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
PUSH="${PUSH:-false}"
SCAN="${SCAN:-true}"

# Print configuration
echo -e "${BLUE}Building Docker image for ${SERVICE_NAME}${NC}"
echo "Registry: ${REGISTRY}"
echo "Version: ${VERSION}"
echo "Git Commit: ${GIT_COMMIT}"
echo "Build Time: ${BUILD_TIME}"
echo "Platforms: ${PLATFORMS}"
echo "Push: ${PUSH}"
echo ""

# Function to build image
build_image() {
    local tag=$1
    local dockerfile=${2:-Dockerfile}

    echo -e "${YELLOW}Building image: ${REGISTRY}/${SERVICE_NAME}:${tag}${NC}"

    # Build arguments
    BUILD_ARGS="--build-arg VERSION=${VERSION}"
    BUILD_ARGS="${BUILD_ARGS} --build-arg BUILD_TIME=${BUILD_TIME}"
    BUILD_ARGS="${BUILD_ARGS} --build-arg GIT_COMMIT=${GIT_COMMIT}"

    # Platform configuration
    if [ "${PLATFORMS}" != "linux/amd64" ]; then
        # Multi-platform build requires buildx
        if ! docker buildx inspect multi-builder > /dev/null 2>&1; then
            echo -e "${YELLOW}Creating buildx builder for multi-platform builds${NC}"
            docker buildx create --name multi-builder --use
        fi
        PLATFORM_ARG="--platform ${PLATFORMS}"
        BUILDER="buildx build"
        if [ "${PUSH}" == "true" ]; then
            PUSH_ARG="--push"
        else
            PUSH_ARG="--load"
        fi
    else
        PLATFORM_ARG=""
        BUILDER="build"
        PUSH_ARG=""
    fi

    # Build command
    docker ${BUILDER} \
        ${PLATFORM_ARG} \
        ${BUILD_ARGS} \
        -f deploy/docker/${dockerfile} \
        -t ${REGISTRY}/${SERVICE_NAME}:${tag} \
        ${PUSH_ARG} \
        .

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Successfully built ${REGISTRY}/${SERVICE_NAME}:${tag}${NC}"
    else
        echo -e "${RED}✗ Failed to build image${NC}"
        exit 1
    fi
}

# Function to scan image
scan_image() {
    local tag=$1

    if [ "${SCAN}" == "true" ] && command -v trivy &> /dev/null; then
        echo -e "${YELLOW}Scanning image for vulnerabilities...${NC}"
        trivy image --severity HIGH,CRITICAL ${REGISTRY}/${SERVICE_NAME}:${tag}

        if [ $? -ne 0 ]; then
            echo -e "${RED}⚠ Security vulnerabilities found!${NC}"
            # Don't exit, just warn
        else
            echo -e "${GREEN}✓ No high/critical vulnerabilities found${NC}"
        fi
    fi
}

# Function to push image
push_image() {
    local tag=$1

    if [ "${PUSH}" == "true" ]; then
        echo -e "${YELLOW}Pushing image: ${REGISTRY}/${SERVICE_NAME}:${tag}${NC}"
        docker push ${REGISTRY}/${SERVICE_NAME}:${tag}

        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Successfully pushed ${REGISTRY}/${SERVICE_NAME}:${tag}${NC}"
        else
            echo -e "${RED}✗ Failed to push image${NC}"
            exit 1
        fi
    fi
}

# Function to tag image
tag_image() {
    local source=$1
    local target=$2

    echo -e "${YELLOW}Tagging ${source} as ${target}${NC}"
    docker tag ${REGISTRY}/${SERVICE_NAME}:${source} ${REGISTRY}/${SERVICE_NAME}:${target}
}

# Main build process
main() {
    # Change to repository root
    cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

    # Build main image
    build_image "${VERSION}"

    # Scan the image
    scan_image "${VERSION}"

    # Additional tags
    if [[ "${VERSION}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        # This is a semantic version tag
        # Also tag as latest
        tag_image "${VERSION}" "latest"
        if [ "${PUSH}" == "true" ]; then
            push_image "latest"
        fi
    elif [ "${VERSION}" == "dev" ] || [[ "${VERSION}" =~ -dirty$ ]]; then
        # Development build
        tag_image "${VERSION}" "dev"
        if [ "${PUSH}" == "true" ] && [ "${VERSION}" != "dev" ]; then
            push_image "dev"
        fi
    fi

    # Build development image if requested
    if [ "${BUILD_DEV}" == "true" ]; then
        build_image "${VERSION}-dev" "Dockerfile.dev"
        if [ "${PUSH}" == "true" ]; then
            push_image "${VERSION}-dev"
        fi
    fi

    echo ""
    echo -e "${GREEN}Build complete!${NC}"
    echo ""
    echo "To run the image locally:"
    echo "  docker run -p 8080:8080 ${REGISTRY}/${SERVICE_NAME}:${VERSION}"
    echo ""
    if [ "${PUSH}" != "true" ]; then
        echo "To push the image:"
        echo "  docker push ${REGISTRY}/${SERVICE_NAME}:${VERSION}"
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --push)
            PUSH="true"
            shift
            ;;
        --no-scan)
            SCAN="false"
            shift
            ;;
        --dev)
            BUILD_DEV="true"
            shift
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --platform)
            PLATFORMS="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --push          Push image to registry"
            echo "  --no-scan       Skip vulnerability scanning"
            echo "  --dev           Also build development image"
            echo "  --version       Override version tag"
            echo "  --registry      Override registry"
            echo "  --platform      Override platforms (default: linux/amd64,linux/arm64)"
            echo "  --help          Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Run main build process
main
