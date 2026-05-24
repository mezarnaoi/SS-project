import React, { useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../../utils/api';
import { useAuth } from '../../contexts/AuthContext';

type Role = 'admin' | 'user';

interface ManagedUser {
  id: string;
  email: string;
  role: Role;
  pages: string[];
}

const PAGE_OPTIONS = ['photos', 'devices', 'statistics', 'reports'] as const;

const defaultPages = ['reports'];

const UsersPage: React.FC = () => {
  const { isAdmin, token } = useAuth();
  const [users, setUsers] = useState<ManagedUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<Role>('user');
  const [pages, setPages] = useState<string[]>(defaultPages);

  const authHeaders = useMemo(
    () => ({
      Authorization: `Bearer ${token ?? ''}`,
      'Content-Type': 'application/json',
    }),
    [token]
  );

  const resetForm = () => {
    setEditingId(null);
    setEmail('');
    setPassword('');
    setRole('user');
    setPages(defaultPages);
  };

  const fetchUsers = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiFetch('/users', {
        headers: authHeaders,
      });
      if (!response.ok) {
        throw new Error(`Failed to fetch users (${response.status})`);
      }
      const data = (await response.json()) as ManagedUser[];
      const normalized = data.map((user) => ({
        ...user,
        pages: Array.isArray(user.pages) && user.pages.length > 0 ? user.pages : ['reports'],
      }));
      setUsers(normalized);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isAdmin && token) {
      void fetchUsers();
    } else {
      setLoading(false);
    }
  }, [isAdmin, token]);

  const togglePagePermission = (page: string, checked: boolean) => {
    if (checked && page !== 'reports') {
      const shouldContinue = window.confirm('Are you sure you want to include this page?');
      if (!shouldContinue) {
        return;
      }
    }

    setPages((current) => {
      if (checked) {
        const next = [...new Set([...current, page])];
        return next;
      }

      const next = current.filter((item) => item !== page);
      return next.length === 0 ? ['reports'] : next;
    });
  };

  const startEdit = (user: ManagedUser) => {
    setEditingId(user.id);
    setEmail(user.email);
    setPassword('');
    setRole(user.role);
    setPages(user.role === 'admin' ? PAGE_OPTIONS.slice() : user.pages);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      const payload: Record<string, unknown> = {
        email,
        role,
        pages: role === 'admin' ? ['photos', 'devices', 'statistics', 'reports', 'users'] : pages,
      };
      if (password.trim()) {
        payload.password = password;
      }

      const endpoint = editingId ? `/users/${editingId}` : '/users';
      const method = editingId ? 'PATCH' : 'POST';
      if (!editingId && !password.trim()) {
        throw new Error('Password is required when creating a user');
      }

      const response = await apiFetch(endpoint, {
        method,
        headers: authHeaders,
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        const message = await response.text();
        throw new Error(message || `Failed to ${editingId ? 'update' : 'create'} user`);
      }

      await fetchUsers();
      resetForm();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Delete this user?')) {
      return;
    }
    setError(null);
    try {
      const response = await apiFetch(`/users/${id}`, {
        method: 'DELETE',
        headers: authHeaders,
      });
      if (!response.ok) {
        throw new Error(`Failed to delete user (${response.status})`);
      }
      await fetchUsers();
      if (editingId === id) {
        resetForm();
      }
    } catch (err) {
      setError((err as Error).message);
    }
  };

  if (!isAdmin) {
    return <div className="max-w-3xl mx-auto py-10">You do not have permission to access this page.</div>;
  }

  return (
    <div className="max-w-5xl mx-auto py-6 space-y-6">
      <h1 className="text-2xl font-semibold text-sky-700 dark:text-sky-400">Users Management</h1>
      <div className="rounded-md border border-amber-300 bg-amber-50 px-4 py-2 text-sm text-amber-900">
        Permission changes apply after next login.
      </div>

      {error && (
        <div className="rounded-md border border-red-300 bg-red-50 px-4 py-2 text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <h2 className="text-lg font-medium mb-3">{editingId ? 'Edit User' : 'Create User'}</h2>
        <form onSubmit={handleSubmit} className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <input
            type="email"
            placeholder="Email"
            className="border rounded-md px-3 py-2 text-gray-900"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            required
            disabled={submitting}
          />
          <input
            type="password"
            placeholder={editingId ? 'New password (optional)' : 'Password'}
            className="border rounded-md px-3 py-2 text-gray-900"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            disabled={submitting}
            required={!editingId}
          />

          <select
            className="border rounded-md px-3 py-2 text-gray-900"
            value={role}
            onChange={(event) => setRole(event.target.value as Role)}
            disabled={submitting}
          >
            <option value="user">User</option>
            <option value="admin">Admin</option>
          </select>

          <div className="md:col-span-2">
            <div className="text-sm font-medium mb-1">Pages</div>
            <div className="flex flex-wrap gap-3">
              {PAGE_OPTIONS.map((page) => (
                <label key={page} className="inline-flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={role === 'admin' ? true : pages.includes(page)}
                    disabled={submitting || role === 'admin'}
                    onChange={(event) => togglePagePermission(page, event.target.checked)}
                  />
                  <span className="capitalize">{page}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="md:col-span-2 flex gap-2">
            <button
              type="submit"
              className="px-4 py-2 rounded-md bg-sky-600 text-white hover:bg-sky-700 disabled:opacity-50"
              disabled={submitting}
            >
              {submitting ? 'Saving...' : editingId ? 'Update User' : 'Create User'}
            </button>
            {editingId && (
              <button
                type="button"
                className="px-4 py-2 rounded-md border border-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                onClick={resetForm}
              >
                Cancel
              </button>
            )}
          </div>
        </form>
      </div>

      <div className="rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-100 dark:bg-gray-800">
            <tr>
              <th className="text-left px-3 py-2">Email</th>
              <th className="text-left px-3 py-2">Password</th>
              <th className="text-left px-3 py-2">Role</th>
              <th className="text-left px-3 py-2">Pages</th>
              <th className="text-left px-3 py-2">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td className="px-3 py-3" colSpan={5}>Loading users...</td>
              </tr>
            ) : users.length === 0 ? (
              <tr>
                <td className="px-3 py-3" colSpan={5}>No users found.</td>
              </tr>
            ) : (
              users.map((user) => {
                const userPages = Array.isArray(user.pages) && user.pages.length > 0 ? user.pages : ['reports'];
                return (
                  <tr key={user.id} className="border-t border-gray-200 dark:border-gray-700">
                    <td className="px-3 py-2">{user.email}</td>
                    <td className="px-3 py-2">********</td>
                    <td className="px-3 py-2 capitalize">{user.role}</td>
                    <td className="px-3 py-2">
                      {user.role === 'admin' ? 'All' : userPages.join(', ')}
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex gap-2">
                        <button
                          className="px-2 py-1 rounded border border-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                          onClick={() => startEdit(user)}
                        >
                          Edit
                        </button>
                        <button
                          className="px-2 py-1 rounded border border-red-300 text-red-600 hover:bg-red-50"
                          onClick={() => handleDelete(user.id)}
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default UsersPage;
