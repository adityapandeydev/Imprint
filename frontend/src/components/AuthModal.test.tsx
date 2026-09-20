import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AuthModal } from './AuthModal';
import * as AuthContextModule from '../context/AuthContext';

describe('AuthModal component', () => {
  const mockLogin = vi.fn();
  const mockRegister = vi.fn();
  const mockOpenAuthModal = vi.fn();
  const mockCloseAuthModal = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  function renderModal(
    isOpen = true,
    mode: 'login' | 'register' = 'login'
  ) {
    vi.spyOn(AuthContextModule, 'useAuth').mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      isAuthModalOpen: isOpen,
      authModalMode: mode,
      openAuthModal: mockOpenAuthModal,
      closeAuthModal: mockCloseAuthModal,
      login: mockLogin,
      register: mockRegister,
      logout: vi.fn(),
      updateUserProfileVisibility: vi.fn(),
    });

    return render(<AuthModal />);
  }

  it('renders nothing when isAuthModalOpen is false', () => {
    renderModal(false);
    expect(screen.queryByText(/Welcome Back/i)).not.toBeInTheDocument();
  });

  it('renders Sign In tab and form when modal is open in login mode', () => {
    renderModal(true, 'login');

    expect(screen.getByText('Welcome Back')).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/reader@domain\.com/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Enter your password/i)).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: /^Sign In$/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: /^Sign In to Imprint$/i })
    ).toBeInTheDocument();
  });

  it('calls openAuthModal("register") when switching to Create Account tab', () => {
    renderModal(true, 'login');

    const registerTab = screen.getByRole('button', { name: /Create Account/i });
    fireEvent.click(registerTab);

    expect(mockOpenAuthModal).toHaveBeenCalledWith('register');
  });

  it('shows error banner when submitting with empty credentials in login mode', async () => {
    renderModal(true, 'login');

    const submitBtn = screen.getByRole('button', { name: /^Sign In to Imprint$/i });
    const form = submitBtn.closest('form')!;
    fireEvent.submit(form);

    await waitFor(() => {
      expect(
        screen.getByText(/Please enter your email or username/i)
      ).toBeInTheDocument();
    });
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('calls login when valid credentials are submitted', async () => {
    renderModal(true, 'login');

    const emailInput = screen.getByPlaceholderText(/reader@domain\.com/i);
    const passwordInput = screen.getByPlaceholderText(/Enter your password/i);

    fireEvent.change(emailInput, { target: { value: 'aditya@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'SecretPassword123' } });

    const submitBtn = screen.getByRole('button', { name: /^Sign In to Imprint$/i });
    const form = submitBtn.closest('form')!;
    fireEvent.submit(form);

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        login: 'aditya@example.com',
        password: 'SecretPassword123',
      });
    });
  });

  it('calls closeAuthModal when close button is clicked', () => {
    renderModal(true, 'login');

    const closeBtn = screen.getByRole('button', { name: /Close dialog/i });
    fireEvent.click(closeBtn);

    expect(mockCloseAuthModal).toHaveBeenCalled();
  });
});
