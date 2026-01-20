import express from 'express';
import { authMiddleware } from './middleware/auth';
import { usersRouter } from './routes/users';
import { productsRouter } from './routes/products';

const app = express();

app.use(express.json());
app.use(authMiddleware);

app.use('/users', usersRouter);
app.use('/products', productsRouter);

app.listen(3000, () => {
  console.log('Simple Shop API running on port 3000');
});
