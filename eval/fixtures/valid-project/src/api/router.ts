import express from 'express';
import { authMiddleware } from './middleware/auth';
import { validationMiddleware } from './middleware/validation';
import { errorHandler } from './middleware/error';
import { userRoutes } from './handlers/user';

// Line 15 - Router class
export class Router {
  private app: express.Application;

  constructor(app: express.Application) {
    console.log('Router initialized');
    this.app = app;
    this.setupRoutes();
  }

  private setupRoutes() {
    // Apply middleware
    this.app.use(express.json());
    this.app.use(authMiddleware);
    this.app.use(validationMiddleware);

    // Register routes
    this.app.use('/api/users', userRoutes);

    // Error handling
    this.app.use(errorHandler);
  }
}

export function createRouter(app: express.Application): Router {
  return new Router(app);
}
