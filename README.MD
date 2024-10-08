# TicTacToe Multiplayer 🎮

TicTacToe Multiplayer is an exciting networked version of the classic game. Enjoy the timeless strategy of Tic-Tac-Toe with friends over the internet!

## 🌟 Features

- **Online Multiplayer**: Play with friends over the internet.
- **Simple Interface**: Intuitive gameplay suitable for all ages.
- **Cross-Platform**: Works on Windows, macOS, and Linux.

## 🛠 Tech Stack

- **Backend**: Go using gRPC
- **Client**: Go using Fyne
- **Deployment**: Docker

## 🚀 Quick Start

1. Clone the repository:
   ```
   git clone https://github.com/kuudori/tictactoe.git
   cd tictactoe
   ```

2. Build and run the server:
   ```
   docker build -t tictactoe-server .
   docker run -p 17077:17077 tictactoe-server
   ```

3. Run the client:
    - Use pre-built clients from GitHub assets, or
    - Build the client using fyne-cross:
      ```
      cd client
      fyne-cross windows -arch=amd64
      fyne-cross darwin -arch=amd64
      fyne-cross linux -arch=amd64
      ```

4. Start playing!

## 🎯 Project Goals

TicTacToe Multiplayer was developed as a practical exercise in creating networked applications with Go. It demonstrates:

- Use of gRPC for network communication
- Creation of cross-platform GUI applications with Fyne
- Containerization with Docker
- Multiplayer game logic implementation

## 🤝 Contributing

We welcome contributions! Please feel free to submit a Pull Request.

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

*Note: This README was generated by AI.*

Enjoy your game of TicTacToe! ⭕❌