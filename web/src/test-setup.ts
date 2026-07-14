// jest-dom matchers (toBeInTheDocument, …) wired into vitest's expect.
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@solidjs/testing-library';
import { afterEach } from 'vitest';

// globals:false means testing-library can't auto-register its afterEach cleanup,
// so do it explicitly — otherwise mounted DOM leaks across tests.
afterEach(() => cleanup());
