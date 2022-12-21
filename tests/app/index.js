const express = require('express');
const app = express();

app.get('/', (req, res) => {
  const name = process.env.NAME || 'Test';
  res.send(`Hello ${name}! (1671659029)`);
});

const port = parseInt(process.env.PORT) || 8080;
app.listen(port, () => {
  console.log(`test 1671659029: listening on port ${port}`);
});
