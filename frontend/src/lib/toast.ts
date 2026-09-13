export interface ToastMessage {
  id: string;
  title: string;
  description?: string;
  type?: 'success' | 'info' | 'error';
}

type ToastListener = (toast: ToastMessage) => void;

class ToastManager {
  private listeners: Set<ToastListener> = new Set();

  subscribe(listener: ToastListener) {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  show(toast: Omit<ToastMessage, 'id'>) {
    const id = Math.random().toString(36).substring(2, 9);
    const item: ToastMessage = { ...toast, id };
    this.listeners.forEach((listener) => listener(item));
  }

  success(title: string, description?: string) {
    this.show({ title, description, type: 'success' });
  }

  info(title: string, description?: string) {
    this.show({ title, description, type: 'info' });
  }

  error(title: string, description?: string) {
    this.show({ title, description, type: 'error' });
  }
}

export const toast = new ToastManager();
