import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import React from 'react';
import { AuthProvider, useAuth } from './AuthContext.jsx';

function Consumer() {
  const { user, saveAuth, logout } = useAuth();
  return (
    <div>
      <div data-testid='username'>{user?.linuxdo_username || 'guest'}</div>
      <button
        type='button'
        onClick={() => saveAuth('token', { linuxdo_username: 'tester', role: 'user' })}
      >
        login
      </button>
      <button type='button' onClick={logout}>
        logout
      </button>
    </div>
  );
}

describe('AuthContext', () => {
  it('saves and clears auth state', () => {
    render(
      <AuthProvider>
        <Consumer />
      </AuthProvider>,
    );

    expect(screen.getByTestId('username').textContent).toBe('guest');

    fireEvent.click(screen.getByText('login'));
    expect(screen.getByTestId('username').textContent).toBe('tester');

    fireEvent.click(screen.getByText('logout'));
    expect(screen.getByTestId('username').textContent).toBe('guest');
  });
});
