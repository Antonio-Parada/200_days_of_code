const express = require('express');
const http = require('http');
const socketIo = require('socket.io');
const session = require('express-session');
const bcrypt = require('bcrypt');
const { pool } = require('./db');

const app = express();
const server = http.createServer(app);
const io = socketIo(server);

const PORT = process.env.PORT || 3000;

// Store connected users: username -> socket.id
const connectedUsers = {};
// Store group members: groupName -> [username1, username2, ...]
const groups = {};

// Middleware for parsing JSON and URL-encoded data
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Session middleware
app.use(session({
  secret: 'your_secret_key', // Replace with a strong secret key
  resave: false,
  saveUninitialized: false,
  cookie: { secure: false } // Set to true if using HTTPS
}));

// Serve static files (like index.html)
app.get('/', (req, res) => {
  res.sendFile(__dirname + '/index.html');
});

// User registration route
app.post('/register', async (req, res) => {
  const { username, password } = req.body;

  if (!username || !password) {
    return res.status(400).send('Username and password are required');
  }

  try {
    const hashedPassword = await bcrypt.hash(password, 10);
    const result = await pool.query(
      'INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id',
      [username, hashedPassword]
    );
    res.status(201).send(`User registered with ID: ${result.rows[0].id}`);
  } catch (err) {
    if (err.code === '23505') { // Unique violation error code
      res.status(409).send('Username already exists');
    } else {
      console.error('Error during registration:', err);
      res.status(500).send('Server error');
    }
  }
});

// User login route
app.post('/login', async (req, res) => {
  const { username, password } = req.body;

  if (!username || !password) {
    return res.status(400).send('Username and password are required');
  }

  try {
    const result = await pool.query('SELECT * FROM users WHERE username = $1', [username]);
    const user = result.rows[0];

    if (!user) {
      return res.status(400).send('Invalid username or password');
    }

    const isMatch = await bcrypt.compare(password, user.password);

    if (!isMatch) {
      return res.status(400).send('Invalid username or password');
    }

    req.session.userId = user.id;
    req.session.username = user.username;
    res.status(200).json({ message: 'Logged in successfully', username: user.username });
  } catch (err) {
    console.error('Error during login:', err);
    res.status(500).send('Server error');
  }
});

// User logout route
app.post('/logout', (req, res) => {
  req.session.destroy(err => {
    if (err) {
      console.error('Error destroying session:', err);
      return res.status(500).send('Could not log out');
    }
    res.status(200).send('Logged out successfully');
  });
});

// Middleware to check if user is authenticated (for protected routes)
function isAuthenticated(req, res, next) {
  if (req.session.userId) {
    next();
  } else {
    res.status(401).send('Unauthorized');
  }
}

// Example of a protected route (you can use this for chat functionality later)
app.get('/dashboard', isAuthenticated, (req, res) => {
  res.send(`Welcome to the dashboard, ${req.session.username}!`);
});

// Socket.IO connection handling
io.on('connection', (socket) => {
  console.log('A user connected');

  // When a user logs in, associate their username with their socket ID
  socket.on('user connected', (username) => {
    connectedUsers[username] = socket.id;
    console.log(`User ${username} connected with socket ID: ${socket.id}`);
    // Optionally, broadcast the list of online users to all clients
    io.emit('online users', Object.keys(connectedUsers));
    // Also send the current list of groups
    socket.emit('available groups', Object.keys(groups));
  });

  socket.on('chat message', (data) => {
    // data is expected to be an object: { username: '...', message: '...' }
    io.emit('chat message', data);
  });

  socket.on('private message', (data) => {
    const { recipient, sender, message } = data;
    const recipientSocketId = connectedUsers[recipient];

    if (recipientSocketId) {
      io.to(recipientSocketId).emit('private message', { sender, message });
      // Optionally, send a confirmation to the sender
      io.to(socket.id).emit('private message confirmation', { recipient, message });
    } else {
      io.to(socket.id).emit('private message error', `User ${recipient} is not online.`);
    }
  });

  socket.on('join group', (groupName, username) => {
    socket.join(groupName);
    if (!groups[groupName]) {
      groups[groupName] = [];
    }
    if (!groups[groupName].includes(username)) {
      groups[groupName].push(username);
    }
    console.log(`${username} joined group: ${groupName}`);
    io.emit('available groups', Object.keys(groups)); // Update all clients about new group
    io.to(groupName).emit('group message', { group: groupName, username: 'System', message: `${username} has joined the group.` });
  });

  socket.on('group message', (data) => {
    const { group, username, message } = data;
    if (groups[group] && groups[group].includes(username)) {
      io.to(group).emit('group message', { group, username, message });
    } else {
      io.to(socket.id).emit('group message error', `You are not a member of group ${group}.`);
    }
  });

  socket.on('disconnect', () => {
    console.log('User disconnected');
    // Remove the disconnected user from the connectedUsers map
    for (const username in connectedUsers) {
      if (connectedUsers[username] === socket.id) {
        delete connectedUsers[username];
        console.log(`User ${username} disconnected.`);
        // Remove user from any groups they were in
        for (const groupName in groups) {
          groups[groupName] = groups[groupName].filter(member => member !== username);
          if (groups[groupName].length === 0) {
            delete groups[groupName]; // Remove empty groups
            io.emit('available groups', Object.keys(groups)); // Update all clients
          }
        }
        break;
      }
    }
    // Optionally, broadcast the updated list of online users
    io.emit('online users', Object.keys(connectedUsers));
  });
});

server.listen(PORT, () => {
  console.log(`Server listening on port ${PORT}`);
});