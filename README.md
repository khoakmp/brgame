# Cloud-Based Platform Game Streaming Service

A scalable, cloud-based **platform for game streaming**, enabling both single-player and multiplayer sessions with low-latency video delivery using **WebRTC**, managed via a central **Coordinator** and distributed **Worker nodes**.

---

## How It Works

### System Components:

* **Web Frontend:**

  * Users initiate game sessions via **WebSocket** requests.
  * Sends `start-game` requests including `game-id` and mode (single or multiplayer).

* **Coordinator:**

  * Manages game session orchestration.
  * Handles both single-player and multiplayer session requests.
  * Maintains **waiting rooms** in-memory for multiplayer matchmaking.
  * Facilitates WebRTC signaling (ICE candidate exchange) between clients and workers via WebSocket.

* **Worker Machines:**

  * Host actual game processes.
  * Stream game video/audio to clients via **WebRTC**.
  * Communicate through the coordinator for signaling.

---

## Workflow Overview

### Single Player Session:

1. User sends `start-game` request with `game-id` over WebSocket.
2. Coordinator selects an available Worker.
3. Worker runs the game.
4. Video/audio stream is sent to the client using **WebRTC**.
5. ICE candidate exchange occurs via WebSocket through the Coordinator.

### Multiplayer Session:

1. User sends `start-game` request with `game-id` and multiplayer flag.
2. Coordinator checks active **waiting rooms**:

   * If a room is available, the user joins.
   * If no room matches, user is placed in a new waiting room with a timeout.
3. Once room criteria met (e.g., enough players):

   * Coordinator selects a Worker.
   * Worker runs the game for all players.
4. WebRTC streams are established to all clients.
5. Signaling (ICE candidates) handled through Coordinator.

---

## Technical Highlights

* **WebSocket Communication:** Real-time signaling and session management.
* **WebRTC Protocol:** Low-latency, peer-to-peer video/audio streaming.
* **In-Memory Matchmaking:** Fast multiplayer session matching with timeouts for unfulfilled rooms.
* **Distributed Workers:** Scalable game processing across multiple nodes.
* **Coordinator Logic:** Central authority for session orchestration, room management, and signaling.

---

## Use Cases

* Cloud Gaming Platforms
* Remote Game Testing Environments
* Multiplayer Game Lobbies with Real-Time Streaming
* Low-Latency Game Demos or Trials

---

## Roadmap

* [ ] Room capacity and customizable matchmaking rules
* [ ] Dynamic scaling of Worker nodes
* [ ] Advanced session timeout and failure handling
* [ ] Player reconnection support
* [ ] Analytics and monitoring dashboards

---

## License

MIT License

---

## Contribution

Contributions welcome! Open issues or submit PRs to enhance scalability, performance, or new features.

---

## Inspiration

Inspired by modern cloud gaming solutions with focus on lightweight signaling, efficient resource management, and seamless multiplayer experiences.
