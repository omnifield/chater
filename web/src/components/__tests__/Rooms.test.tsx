import { fireEvent, render, screen, waitFor } from '@solidjs/testing-library';
import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '../../api/client';
import { Rooms } from '../Rooms';
import { stubApi } from './testApi';

describe('Rooms', () => {
  it('renders the room list', async () => {
    const api = stubApi({
      listRooms: async () => [
        { id: 1, type: 'group', title: 'devs', created_at: 'x' },
        { id: 2, type: 'dialog', title: null, created_at: 'x' },
      ],
    });

    render(() => <Rooms api={api} selectedId={null} onSelect={() => {}} />);

    expect(await screen.findByText('devs')).toBeInTheDocument();
    expect(await screen.findByText('dialog #2')).toBeInTheDocument();
  });

  it('creates a room and selects it', async () => {
    const createRoom = vi.fn(async () => ({
      id: 9,
      type: 'group' as const,
      title: 'new',
      created_at: 'x',
    }));
    const onSelect = vi.fn();
    const api = stubApi({ createRoom });

    render(() => <Rooms api={api} selectedId={null} onSelect={onSelect} />);

    fireEvent.click(await screen.findByText('Create'));

    await waitFor(() => expect(createRoom).toHaveBeenCalledWith('group', undefined));
    await waitFor(() => expect(onSelect).toHaveBeenCalledWith(9));
  });

  it('shows a recoverable error and reloads the list via Retry', async () => {
    let fail = true;
    const api = stubApi({
      listRooms: async () => {
        if (fail) {
          fail = false;
          throw new ApiError(0, 'network error');
        }
        return [{ id: 3, type: 'group' as const, title: 'back', created_at: 'x' }];
      },
    });

    render(() => <Rooms api={api} selectedId={null} onSelect={() => {}} />);

    expect(await screen.findByText(/Could not load rooms/)).toBeInTheDocument();

    fireEvent.click(screen.getByText('Retry'));

    expect(await screen.findByText('back')).toBeInTheDocument();
  });

  it('Create still issues the call even after the initial load failed', async () => {
    const createRoom = vi.fn(async () => ({
      id: 9,
      type: 'group' as const,
      title: null,
      created_at: 'x',
    }));
    const api = stubApi({
      listRooms: async () => {
        throw new ApiError(0, 'network error');
      },
      createRoom,
    });

    render(() => <Rooms api={api} selectedId={null} onSelect={() => {}} />);

    await screen.findByText(/Could not load rooms/);
    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createRoom).toHaveBeenCalledWith('group', undefined));
  });
});
