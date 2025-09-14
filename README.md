# Real-Time Polling Application

This project is a real-time polling application built with Go and a microservices architecture. It demonstrates the use of RabbitMQ for asynchronous communication between services and Redis for data persistence.

## Architecture

The application consists of two main services, a message broker, and a database:

-   **`polling-service`**: A Go service that exposes an HTTP endpoint to cast votes. When a vote is received, it publishes a message to a RabbitMQ queue.
-   **`results-service`**: A Go service that consumes vote messages from the RabbitMQ queue, stores the results in a Redis database, and broadcasts the updated results to connected clients in real-time using WebSockets.
-   **`rabbitmq`**: A RabbitMQ message broker that decouples the `polling-service` from the `results-service`.
-   **`redis`**: A Redis database used to store and persist the vote counts.

## Services

| Service           | Port (Host) | Description                                                                                                                             |
| ----------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `polling-service` | `8080`      | Handles incoming votes and publishes them to the message queue.                                                                         |
| `results-service` | `8081`      | Consumes votes, updates the database, and broadcasts results via WebSockets. Also provides an HTTP endpoint to get the current results. |
| `rabbitmq`        | `15672`     | RabbitMQ management interface.                                                                                                          |
| `redis`           | `6379`      | Redis database port.                                                                                                                    |

## Getting Started

To run the application, you need to have Docker and Docker Compose installed.

1.  Clone the repository:

    ```bash
    git clone https://github.com/lekob/pub-sub-playground.git
    cd pub-sub-playground
    ```

2.  Build and run the services using Docker Compose:

    ```bash
    docker compose up --build -d
    ```

## Usage

### Casting a Vote

To cast a vote, send a `POST` request to the `polling-service`.

**Endpoint**: `http://localhost:8080/vote`

**Method**: `POST`

**Body**:

```json
{
    "option": "go"
}
```

You can use `curl` to cast a vote:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"option": "go"}' http://localhost:8080/vote
```

### Viewing Results

There are three ways to view the results:

1. **HTTP Endpoint**: Get the current results by sending a `GET` request to the `results-service`.

    **Endpoint**: `http://localhost:8081/results`

    **Method**: `GET`

    Example response:

    ```json
    {
        "go": 10,
        "python": 5,
        "javascript": 8
    }
    ```

2. **Real-Time Updates via WebSocket**: Connect to the WebSocket endpoint to receive real-time updates as votes are cast.

    **Endpoint**: `ws://localhost:8081/ws`

    You can use a WebSocket client like `websocat` to connect:

    ```bash
    websocat ws://localhost:8081/ws
    ```

    Each time a vote is cast, you will receive a message with the updated vote counts.

3. **Web Interface**: Open the `index.html` file in your browser to view the results in real-time with a user-friendly interface.

    - The web interface connects to the WebSocket endpoint (`ws://localhost:8081/ws`) to display real-time updates.
    - To use the interface, simply open the `index.html` file in your browser:

      ```bash
      xdg-open index.html
      ```

    - The interface will automatically connect to the WebSocket server and display the poll results dynamically.
