import { ws } from 'msw';
import { db, marketMovement, toFrame } from './state';

export const availabilityLink = ws.link('*/ws/events/:id');

const MARKET_TICK_MS = 8_000;

export const wsHandlers = [
  availabilityLink.addEventListener('connection', ({ client, params }) => {
    const event = db.events.find((e) => e.id === params.id);
    if (!event) {
      client.close();
      return;
    }

    client.send(JSON.stringify(toFrame(event)));

    const timer = setInterval(() => {
      marketMovement(event.id);
      client.send(JSON.stringify(toFrame(event)));
    }, MARKET_TICK_MS);

    client.addEventListener('close', () => clearInterval(timer));
  }),
];
