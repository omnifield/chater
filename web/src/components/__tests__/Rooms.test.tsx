import { fireEvent, render, screen, waitFor } from '@solidjs/testing-library';
import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '../../api/client';
import type { Room } from '../../api/types';
import { mergeRooms, Rooms } from '../Rooms';
import { stubApi } from './testApi';

const room = (id: number, title: string | null): Room => ({
  id,
  type: 'group',
  title,
  created_at: `t${id}`,
});

describe('mergeRooms', () => {
  it('reuses unchanged rooms, adds new ones, drops gone ones', () => {
    const a = room(1, 'a');
    const b = room(2, 'b');
    const merged = mergeRooms([a, b], [{ ...a }, room(3, 'c')]);

    expect(merged[0]).toBe(a); // same reference -> no re-render
    expect(merged.map((r) => r.id)).toEqual([1, 3]); // 2 dropped, 3 added
  });

  it('replaces a room whose title changed', () => {
    const a = room(1, 'old');
    const merged = mergeRooms([a], [room(1, 'new')]);
    expect(merged[0]).not.toBe(a);
    expect(merged[0].title).toBe('new');
  });
});

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

  it('polls and surfaces a newly-added room without a manual refetch', async () => {
    let listing: Room[] = [room(1, 'one')];
    const api = stubApi({ listRooms: async () => listing.slice() });

    render(() => <Rooms api={api} selectedId={1} onSelect={() => {}} pollMs={20} />);

    expect(await screen.findByText('one')).toBeInTheDocument();

    // someone adds us to a new room; the next poll must surface it
    listing = [room(2, 'two'), ...listing];

    expect(await screen.findByText('two')).toBeInTheDocument();
    expect(screen.getByText('one')).toBeInTheDocument(); // existing room kept
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
