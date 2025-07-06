const express = require('express');
const axios = require('axios');

const app = express();
const PORT = process.env.PORT || 3001; // Node.js backend will run on port 3001

// Middleware to parse JSON bodies
app.use(express.json());

// Define the URL for the Python FastAPI backend
const PYTHON_BACKEND_URL = process.env.PYTHON_BACKEND_URL || 'http://localhost:8000';

// Example route to fetch products from the Python backend
app.get('/api/products', async (req, res) => {
  try {
    const response = await axios.get(`${PYTHON_BACKEND_URL}/products`);
    res.json(response.data);
  } catch (error) {
    console.error('Error fetching products from Python backend:', error.message);
    res.status(500).json({ message: 'Error fetching products', error: error.message });
  }
});

// Basic route for testing
app.get('/', (req, res) => {
  res.send('Node.js API Gateway is running!');
});

app.listen(PORT, () => {
  console.log(`Node.js API Gateway listening on port ${PORT}`);
});