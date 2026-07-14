import type { ChatApi } from '../../api/client';

/** stubApi builds a ChatApi with harmless defaults, overridable per test. */
export function stubApi(over: Partial<ChatApi> = {}): ChatApi {
  return {
    listRooms: async () => [],
    createRoom: async () => ({ id: 0, type: 'group', title: null, created_at: '' }),
    addParticipant: async () => {},
    getMessages: async () => ({ messages: [], next_cursor: null }),
    sendMessage: async () => ({ id: 0, room_id: 0, author_id: 0, body: '', created_at: '' }),
    openRoomSocket: () => ({ close: () => {} }),
    ...over,
  };
}
