#!/bin/bash
# Local development environment management script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default values
ACTION=""
VERBOSE=false
NO_LOGS=false
SERVICES="postgres redis otel-collector prometheus grafana jaeger"

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Function to print header
print_header() {
    echo ""
    print_color "$BLUE" "========================================="
    print_color "$BLUE" "  lumi-go Local Development Environment"
    print_color "$BLUE" "========================================="
    echo ""
}

# Function to check dependencies
check_dependencies() {
    local missing_deps=()

    # Check for Docker
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi

    # Check for Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        missing_deps+=("docker-compose")
    fi

    # Check for Make
    if ! command -v make &> /dev/null; then
        missing_deps+=("make")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_color "$RED" "✗ Missing dependencies: ${missing_deps[*]}"
        print_color "$YELLOW" "Please install the missing dependencies and try again."
        exit 1
    fi

    print_color "$GREEN" "✓ All dependencies are installed"
}

# Function to check Docker daemon
check_docker_daemon() {
    if ! docker info &> /dev/null; then
        print_color "$RED" "✗ Docker daemon is not running"
        print_color "$YELLOW" "Please start Docker and try again."
        exit 1
    fi
    print_color "$GREEN" "✓ Docker daemon is running"
}

# Function to start services
start_services() {
    print_color "$YELLOW" "Starting services..."

    cd "$PROJECT_ROOT"

    # Start docker-compose services
    docker-compose up -d

    # Wait for services to be healthy
    print_color "$YELLOW" "Waiting for services to be healthy..."

    # Wait for PostgreSQL
    print_color "$CYAN" "  • Waiting for PostgreSQL..."
    until docker-compose exec -T postgres pg_isready -U lumigo &> /dev/null; do
        sleep 1
    done
    print_color "$GREEN" "  ✓ PostgreSQL is ready"

    # Wait for Redis
    print_color "$CYAN" "  • Waiting for Redis..."
    until docker-compose exec -T redis redis-cli ping &> /dev/null; do
        sleep 1
    done
    print_color "$GREEN" "  ✓ Redis is ready"

    # Run migrations
    print_color "$YELLOW" "Running database migrations..."
    docker-compose run --rm migrate
    print_color "$GREEN" "✓ Migrations complete"

    # Seed database (if seed script exists)
    if [ -f "$PROJECT_ROOT/scripts/seed.sql" ]; then
        print_color "$YELLOW" "Seeding database..."
        docker-compose exec -T postgres psql -U lumigo -d lumigo < "$PROJECT_ROOT/scripts/seed.sql"
        print_color "$GREEN" "✓ Database seeded"
    fi

    print_color "$GREEN" "✓ All services started successfully!"
    echo ""
    print_color "$BLUE" "Service URLs:"
    print_color "$CYAN" "  • Application:    http://localhost:8080"
    print_color "$CYAN" "  • Metrics:        http://localhost:9090/metrics"
    print_color "$CYAN" "  • Prometheus:     http://localhost:9091"
    print_color "$CYAN" "  • Grafana:        http://localhost:3000 (admin/admin)"
    print_color "$CYAN" "  • Jaeger:         http://localhost:16686"
    print_color "$CYAN" "  • PostgreSQL:     localhost:5432 (lumigo/lumigo)"
    print_color "$CYAN" "  • Redis:          localhost:6379"
    echo ""

    if [ "$NO_LOGS" = false ]; then
        print_color "$YELLOW" "Showing logs (press Ctrl+C to exit)..."
        docker-compose logs -f app
    fi
}

# Function to stop services
stop_services() {
    print_color "$YELLOW" "Stopping services..."

    cd "$PROJECT_ROOT"
    docker-compose down

    print_color "$GREEN" "✓ All services stopped"
}

# Function to restart services
restart_services() {
    stop_services
    start_services
}

# Function to show status
show_status() {
    print_color "$YELLOW" "Service Status:"
    echo ""

    cd "$PROJECT_ROOT"
    docker-compose ps

    echo ""
    print_color "$YELLOW" "Container Resource Usage:"
    docker stats --no-stream $(docker-compose ps -q) 2>/dev/null || true
}

# Function to show logs
show_logs() {
    local service=$1

    cd "$PROJECT_ROOT"

    if [ -n "$service" ]; then
        print_color "$YELLOW" "Showing logs for $service..."
        docker-compose logs -f "$service"
    else
        print_color "$YELLOW" "Showing logs for all services..."
        docker-compose logs -f
    fi
}

# Function to clean up
cleanup() {
    print_color "$YELLOW" "Cleaning up..."

    cd "$PROJECT_ROOT"

    # Stop and remove containers, networks, volumes
    docker-compose down -v

    # Remove temporary files
    rm -rf "$PROJECT_ROOT/tmp"
    rm -f "$PROJECT_ROOT/air.log"

    print_color "$GREEN" "✓ Cleanup complete"
}

# Function to run database console
db_console() {
    print_color "$YELLOW" "Connecting to PostgreSQL..."
    cd "$PROJECT_ROOT"
    docker-compose exec postgres psql -U lumigo -d lumigo
}

# Function to run Redis console
redis_console() {
    print_color "$YELLOW" "Connecting to Redis..."
    cd "$PROJECT_ROOT"
    docker-compose exec redis redis-cli
}

# Function to reset database
reset_db() {
    print_color "$YELLOW" "Resetting database..."

    cd "$PROJECT_ROOT"

    # Run down migrations
    docker-compose run --rm migrate -path /migrations -database "postgres://lumigo:lumigo@postgres:5432/lumigo?sslmode=disable" down -all

    # Run up migrations
    docker-compose run --rm migrate -path /migrations -database "postgres://lumigo:lumigo@postgres:5432/lumigo?sslmode=disable" up

    # Seed database if seed script exists
    if [ -f "$PROJECT_ROOT/scripts/seed.sql" ]; then
        print_color "$YELLOW" "Seeding database..."
        docker-compose exec -T postgres psql -U lumigo -d lumigo < "$PROJECT_ROOT/scripts/seed.sql"
    fi

    print_color "$GREEN" "✓ Database reset complete"
}

# Function to show help
show_help() {
    print_header

    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  start         Start all services"
    echo "  stop          Stop all services"
    echo "  restart       Restart all services"
    echo "  status        Show service status"
    echo "  logs [name]   Show logs (optionally for specific service)"
    echo "  clean         Stop services and clean up volumes"
    echo "  db            Open PostgreSQL console"
    echo "  redis         Open Redis console"
    echo "  reset-db      Reset database (drop and recreate)"
    echo "  help          Show this help message"
    echo ""
    echo "Options:"
    echo "  -v, --verbose    Show verbose output"
    echo "  -n, --no-logs    Don't show logs after starting"
    echo ""
    echo "Examples:"
    echo "  $0 start              # Start all services"
    echo "  $0 start --no-logs    # Start without showing logs"
    echo "  $0 logs app           # Show logs for app service"
    echo "  $0 status             # Show status of all services"
    echo "  $0 clean              # Stop and clean up everything"
    echo ""
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        start|stop|restart|status|logs|clean|db|redis|reset-db|help)
            ACTION=$1
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -n|--no-logs)
            NO_LOGS=true
            shift
            ;;
        *)
            # For logs command, this could be the service name
            if [ "$ACTION" = "logs" ]; then
                SERVICE_NAME=$1
            else
                print_color "$RED" "Unknown option: $1"
                show_help
                exit 1
            fi
            shift
            ;;
    esac
done

# If no action specified, show help
if [ -z "$ACTION" ]; then
    show_help
    exit 0
fi

# Execute action
print_header

case $ACTION in
    start)
        check_dependencies
        check_docker_daemon
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        check_dependencies
        check_docker_daemon
        restart_services
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$SERVICE_NAME"
        ;;
    clean)
        cleanup
        ;;
    db)
        db_console
        ;;
    redis)
        redis_console
        ;;
    reset-db)
        reset_db
        ;;
    help)
        show_help
        ;;
    *)
        print_color "$RED" "Unknown command: $ACTION"
        show_help
        exit 1
        ;;
esac
