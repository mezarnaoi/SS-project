import React, { useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../../utils/api';
import { useAuth } from '../../contexts/AuthContext';

type UserRole = 'admin' | 'user';
type PagePermission = 'photos' | 'devices' | 'statistics' | 'reports';

interface UserRecord {
  id: string;
  email: string;
  password: string;
  role: UserRole;
  pages: string[];
}

interface UserFormState {
  email: string;
  password: string;
  role: UserRole;
  pages: PagePermission[];
}

const ALL_PAGES: PagePermission[] = ['photos', 'devices', 'statistics', 'reports'];

const defaultForm: UserFormState = {
  email: '',
  password: '',
  role: 'user',
  pages: ['reports'],
};

const UsersPage: React.FC = () => {
  const { token, isAdmin } = useAuth();
  const [users, setUsers] = useState<UserRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editingUserId, setEditingUserId] = useState<string | null>(null);
  const [form, setForm] = useState<UserFormState>(defaultForm);

  const isEditMode = useMemo(() => editingUserId !== null, [editingUserId]);

  const fetchUsers = async () => {
    if (!token) return;
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch('/users', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      if (!res.ok) throw new Error('Failed to fetch users');
      const data = await res.json();
      setUsers(Array.isArray(data) ? data : []);
    } catch (err) {
      setError((err as Error).message || 'Failed to fetch users');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const togglePagePermission = (page: PagePermission) => {
    const enabled = form.pages.includes(page);
    if (!enabled && page !== 'reports') {
      const proceed = window.confirm(`Are you sure you want to include this page? (${page})`);
      if (!proceed) return;
    }

    const nextPages = enabled
      ? form.pages.filter((p) => p !== page)
      : [...form.pages, page];

    setForm((prev) => ({
      ...prev,
      pages: nextPages.length === 0 ? ['reports'] : nextPages,
    }));
  };

  const startEdit = (user: UserRecord) => {
    setEditingUserId(user.id);
    setForm({
      email: user.email,
      password: '',
      role: user.role,
      pages: user.role === 'admin'
        ? ['photos', 'devices', 'statistics', 'reports']
        : (user.pages.filter((p): p is PagePermission => ALL_PAGES.includes(p as PagePermission)) as PagePermission[]),
    });
  };

  const resetForm = () => {
    setEditingUserId(null);
    setForm(defaultForm);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;

    try {
      const payload = {
        email: form.email,
        password: form.password,
        role: form.role,
        pages: form.role === 'admin' ? ALL_PAGES : form.pages,
      };

      const res = await apiFetch(isEditMode ? `/users/${editingUserId}` : '/users', {
        method: isEditMode ? 'PATCH' : 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || 'Failed to save user');
      }

      resetForm();
      fetchUsers();
    } catch (err) {
      alert((err as Error).message || 'Failed to save user');
    }
  };

  const handleDelete = async (user: UserRecord) => {
    if (!token) return;
    const confirmDelete = window.confirm(`Delete user ${user.email}?`);
    if (!confirmDelete) return;

    try {
      const res = await apiFetch(`/users/${user.id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || 'Failed to delete user');
      }
      fetchUsers();
    } catch (err) {
      alert((err as Error).message || 'Failed to delete user');
    }
  };

  if (!isAdmin) {
    return (
      <div className="container mx-auto">
        <h1 className="text-2xl font-semibold text-sky-700 mb-6">Users</h1>
        <div className="bg-red-50 border border-red-200 text-red-700 p-4 rounded-md">
          You do not have permission to access this page.
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto">
      <h1 className="text-2xl font-semibold text-sky-700 mb-6">Users</h1>

      <div className="bg-white text-gray-900 p-4 rounded-lg shadow-sm mb-6">
        <h2 className="text-lg font-medium mb-4">{isEditMode ? 'Edit User' : 'Create User'}</h2>
        <form onSubmit={handleSubmit} className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">
              Password {isEditMode ? '(leave empty to keep existing)' : ''}
            </label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              required={!isEditMode}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Role</label>
            <select
              value={form.role}
              onChange={(e) => setForm((prev) => ({ ...prev, role: e.target.value as UserRole }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            >
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Pages</label>
            <div className="flex flex-wrap gap-3 pt-2">
              {ALL_PAGES.map((page) => (
                <label key={page} className="inline-flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={form.role === 'admin' ? true : form.pages.includes(page)}
                    disabled={form.role === 'admin' || (page === 'reports' && form.pages.length === 1 && form.pages.includes('reports'))}
                    onChange={() => togglePagePermission(page)}
                  />
                  {page}
                </label>
              ))}
            </div>
            <p className="text-xs text-gray-500 mt-2">Default access is Reports.</p>
          </div>

          <div className="md:col-span-2 flex gap-3">
            <button
              type="submit"
              className="px-4 py-2 bg-sky-600 text-white rounded-md hover:bg-sky-700"
            >
              {isEditMode ? 'Update User' : 'Create User'}
            </button>
            {isEditMode && (
              <button
                type="button"
                onClick={resetForm}
                className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
              >
                Cancel
              </button>
            )}
          </div>
        </form>
      </div>

      <div className="bg-white text-gray-900 p-4 rounded-lg shadow-sm overflow-x-auto">
        {loading && <p className="text-gray-500">Loading users...</p>}
        {!loading && error && <p className="text-red-600">{error}</p>}
        {!loading && !error && (
          <table className="w-full text-sm">
            <thead className="text-left border-b">
              <tr>
                <th className="py-2 pr-4">Email</th>
                <th className="py-2 pr-4">Password</th>
                <th className="py-2 pr-4">Role</th>
                <th className="py-2 pr-4">Pages</th>
                <th className="py-2 pr-4">Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => {
                const userPages = Array.isArray(user.pages) ? user.pages : ['reports'];
                return (
                  <tr key={user.id} className="border-b last:border-b-0">
                    <td className="py-2 pr-4">{user.email}</td>
                    <td className="py-2 pr-4">******</td>
                    <td className="py-2 pr-4 uppercase">{user.role}</td>
                    <td className="py-2 pr-4">
                      {user.role === 'admin' || userPages.includes('all') ? 'All' : userPages.join(', ')}
                    </td>
                    <td className="py-2 pr-4 flex gap-2">
                      <button
                        onClick={() => startEdit(user)}
                        className="px-3 py-1 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(user)}
                        className="px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default UsersPage;
