import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vitest/config';

const dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
    test: {
        environment: 'jsdom',
        include: ['src/**/*.test.{ts,tsx}'],
    },
    resolve: {
        alias: {
            '@': path.resolve(dirname, './src'),
        },
    },
});