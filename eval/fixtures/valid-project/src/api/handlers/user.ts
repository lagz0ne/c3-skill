import { Router, Request, Response } from 'express';
import { UserModel } from '../models/user';
import { z } from 'zod';

const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1),
});

// Line 10 - UserHandler class
export class UserHandler {
  private model: UserModel;

  constructor() {
    this.model = new UserModel();
  }

  async getAll(req: Request, res: Response) {
    const users = await this.model.findAll();
    res.json(users);
  }

  // Line 25 - createUser function
  async createUser(req: Request, res: Response) {
    const data = createUserSchema.parse(req.body);
    const user = await this.model.create(data);
    res.status(201).json(user);
  }

  async getById(req: Request, res: Response) {
    const user = await this.model.findById(req.params.id);
    if (!user) {
      res.status(404).json({ error: 'User not found' });
      return;
    }
    res.json(user);
  }

  async deleteUser(req: Request, res: Response) {
    const deleted = await this.model.delete(req.params.id);
    if (!deleted) {
      res.status(404).json({ error: 'User not found' });
      return;
    }
    res.status(204).send();
  }
}

const handler = new UserHandler();
export const userRoutes = Router()
  .get('/', handler.getAll.bind(handler))
  .post('/', handler.createUser.bind(handler))
  .get('/:id', handler.getById.bind(handler));
