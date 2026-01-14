import { useState, useEffect } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Shield, Server } from 'lucide-react';
import { toast } from 'sonner';
import { api } from '@/api/client';
import type { ProxyConfig, ApiResponse } from '@/api/types';

export function ProxySettings() {
  const [formData, setFormData] = useState<ProxyConfig>({
    type: 'socks5',
    address: '127.0.0.1:7891',
    username: '',
    password: '',
  });

  const proxyResult = useQuery<ApiResponse<ProxyConfig>>({
    queryKey: ['proxy'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<ProxyConfig>>('/proxy');
      return res.data;
    },
  });

  useEffect(() => {
    if (proxyResult.data?.data) {
      const config = proxyResult.data.data;
      setFormData(prev => ({
        ...prev,
        type: (config.type || 'socks5') as any,
        address: config.address || '',
      }));
    }
  }, [proxyResult.data]);

  const saveProxyMutation = useMutation({
    mutationFn: async (config: ProxyConfig) => {
      await api.post('/proxy', config);
    },
    onSuccess: () => {
      toast.success('Proxy configuration saved successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to update proxy configuration');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveProxyMutation.mutate(formData);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  if (proxyResult.isLoading) {
    return <div className="animate-pulse h-48 bg-gray-800 rounded-xl" />;
  }

  return (
    <div className="rounded-xl bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between pb-4 mb-4">
        <div>
          <h3 className="font-semibold text-gray-900">Proxy Configuration</h3>
          <p className="text-sm text-gray-600 mt-1">Configure SOCKS/HTTP Proxy</p>
        </div>
        <Server className="w-5 h-5 text-gray-600" />
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Protocol Type */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Protocol Type
          </label>
          <select
            name="type"
            value={formData.type}
            onChange={handleChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="socks5">SOCKS5</option>
            <option value="socks4">SOCKS4</option>
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
          </select>
        </div>

        {/* Proxy Address */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Proxy Address
          </label>
          <div className="relative">
            <input
              name="address"
              type="text"
              value={formData.address}
              onChange={handleChange}
              placeholder="127.0.0.1:7891"
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
            />
            <Server className="absolute left-3 top-2.5 w-4 h-4 text-gray-400" />
          </div>
          <p className="mt-1 text-xs text-gray-500">
            Format: host:port (e.g., 127.0.0.1:7891)
          </p>
        </div>

        {/* Username */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Username <span className="text-gray-400 font-normal">(optional)</span>
          </label>
          <div className="relative">
            <input
              name="username"
              type="text"
              value={formData.username}
              onChange={handleChange}
              placeholder="Enter username if required"
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <Shield className="absolute left-3 top-2.5 w-4 h-4 text-gray-400" />
          </div>
        </div>

        {/* Password */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Password <span className="text-gray-400 font-normal">(optional)</span>
          </label>
          <input
            name="password"
            type="password"
            value={formData.password}
            onChange={handleChange}
            placeholder="Enter password if required"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        {/* Save Button */}
        <div className="pt-2">
          <button
            type="submit"
            disabled={saveProxyMutation.isPending}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {saveProxyMutation.isPending ? 'Saving...' : 'Save Configuration'}
          </button>
        </div>
      </form>
    </div>
  );
}
