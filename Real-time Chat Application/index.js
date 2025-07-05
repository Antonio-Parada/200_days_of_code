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

  socket.on('chat message', (data) => {
    // data is expected to be an object: { username: '...', message: '...' }
    io.emit('chat message', data);
  });

  socket.on('disconnect', () => {
    console.log('User disconnected');
  });
});

server.listen(PORT, () => {
  console.log(`Server listening on port ${PORT}`);
});