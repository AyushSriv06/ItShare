# 🔗 ItShare - P2P File Sharing Application 🔗

A peer-to-peer file sharing application with room-based communication and integrated chat functionality, allowing users to create rooms, communicate, and share files directly with each other in organized groups.

> ⚠️ **Note**: *Room-based architecture is currently under development. Some features described below may be part of the upcoming release.*

## ✨ Features

* **👤 User Authentication**: Connect with a username and maintain persistent sessions
* **🏠 Room Management** *(under development)*: Create and join rooms for organized communication
* **💬 Real-time Chat**: Send and receive messages globally or within specific rooms
* **📁 File Sharing**: Transfer files directly between users
* **📂 Folder Sharing**: Share entire folders with other users
* **🔍 File Discovery**: Look up and browse other users' shared directories
* **🎯 Room-based Operations** *(under development)*: File transfers and lookups work within room context
* **🔄 Automatic Reconnection**: Seamlessly reconnect with your existing session
* **👥 Status Tracking**: Monitor which users are currently online
* **🎨 Colorful UI**: Enhanced CLI interface with colors and emojis
* **📊 Progress Bars**: Visual feedback for file and folder transfers
* **🔒 Data Integrity**: MD5 checksum verification for files and folders

## 🚀 Installation

### Prerequisites

* Go (1.16 or later) 🔧

### Steps

1. Clone the repository ⬇️

```bash
git clone https://github.com/AyushSriv06/ItShare.git
cd ItShare
```

2. Build the application 🛠️

```bash
go build -o ItShare
```

## 🎮 Usage

### Starting the Server 🖥️

```bash
# Start server on default port 8080
go run ./server/cmd --port 8080

# Start server on custom port
go run ./server/cmd --port 3000
```

### Connecting as a Client 📱

```bash
# Connect to local server with default port
go run ./client/cmd --server localhost:8080

# Connect to remote server
go run ./client/cmd --server 192.168.0.203:4000
```

The application will validate:

* Server availability before client connection attempts
* Port availability before starting a server
* Existence of shared folder paths

## 🏠 Room-Based Architecture *(under development)*

> Room features are actively being developed and integrated. Below is a preview of the expected workflow and design.

### How Rooms Will Work

1. **🌐 Global Discovery**: All connected users are visible via `/status` command
2. **🏠 Room Creation**: Any user can create a room and invite specific users
3. **💬 Room Chat**: Messages sent within a room are only visible to room participants
4. **📁 Room File Sharing**: File operations (send, lookup, download) work within room context
5. **🎯 Selective Communication**: Users can switch between rooms or communicate globally

## 📝 Commands

### Chat Commands 💬

| Command   | Description                         |
| --------- | ----------------------------------- |
| `/help`   | Show all available commands         |
| `/status` | Show online users                   |
| `exit`    | Disconnect and exit the application |

### Room Management 🏠 *(under development)*

| Command                                          | Description                               |
| ------------------------------------------------ | ----------------------------------------- |
| `/createroom <roomName> <userId1> [userId2] ...` | Create a new room with participants       |
| `/joinroom <roomId>`                             | Join an existing room                     |
| `/leaveroom <roomId>`                            | Leave a room                              |
| `/selectroom <roomId>`                           | Select active room for chat and transfers |
| `/listrooms`                                     | List all available rooms                  |
| `/roominfo <roomId>`                             | Show detailed room information            |

### File Operations 📂

| Command                             | Description                       |
| ----------------------------------- | --------------------------------- |
| `/lookup <userId>`                  | Browse user's shared files        |
| `/sendfile <userId> <filePath>`     | Send a file to another user       |
| `/sendfolder <userId> <folderPath>` | Send a folder to another user     |
| `/download <userId> <filename>`     | Download a file from another user |

> ⚠️ *File operations will work within the context of rooms once the room-based system is live.*

### Transfer Controls 🛁

| Command                | Description               |
| ---------------------- | ------------------------- |
| `/transfers`           | Show all active transfers |
| `/pause <transferId>`  | Pause an active transfer  |
| `/resume <transferId>` | Resume a paused transfer  |

## Terminal UI Features 🎨

* 🌈 **Color-coded messages**:

  * Commands appear in blue
  * Success messages appear in green
  * Error messages appear in red
  * User status notifications in yellow
  * Room messages have special formatting *(under development)*

* 📊 **Progress bars for file transfers**:

  ```
  [===================================>------] 75% (1.2 MB/1.7 MB)
  ```

* 📁 **Improved file listings**:

  ```
  === FOLDERS ===
  📁 [FOLDER] documents (Size: 0 bytes)
  📁 [FOLDER] images (Size: 0 bytes)

  === FILES ===
  📄 [FILE] document.pdf (Size: 1024 bytes)
  📄 [FILE] image.jpg (Size: 2048 bytes)
  ```

* 🏠 **Room indicators** *(planned)*:

  ```
  [Room: MyRoom] >>> Hello everyone in this room!
  ```

## 🎯 Usage Examples *(for upcoming room support)*

### Creating and Using Rooms

```bash
/status
/createroom ProjectTeam 1234 5678
/selectroom 1
Hello team! Let's share some files.
/sendfile 1234 /path/to/document.pdf
/listrooms
/roominfo 1
```

### File Sharing Workflow

```bash
/createroom FileShare 2345
/selectroom 1
/lookup 2345
/sendfile 2345 /path/to/file.txt
/sendfolder 2345 /path/to/folder
/download 2345 filename.txt
```

## 🔒 Security

The application implements several layers of security:

* **📁 Folder Path Validation**
* **🔌 Server Availability Check**
* **🚫 Port Conflict Prevention**
* **🏠 Room-based Access Control** *(in progress)*
* **👥 Session Management**
* **🔐 Checksum Verification**

  Files and folders are transferred with MD5 checksum verification to ensure accuracy. If a mismatch occurs, the user is notified.

---

**Made with ❤️ by the ItShare Team**
