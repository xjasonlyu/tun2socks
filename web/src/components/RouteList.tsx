import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Trash2, Plus, RefreshCw, X } from 'lucide-react';
import { toast } from 'sonner';
import { api } from '@/api/client';
import type { Route, DeviceStatus, ApiResponse } from '@/api/types';

export function RouteList() {
  const queryClient = useQueryClient();
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [formData, setFormData] = useState({
    cidr: '',
    gateway: '198.18.0.1',
    metric: 100,
  });

  const { data: routes, isLoading, refetch } = useQuery({
    queryKey: ['routes'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<Route[]>>('/routes');
      return res.data.data;
    },
  });

  const { data: deviceStatus } = useQuery({
    queryKey: ['device'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<DeviceStatus>>('/device');
      return res.data.data;
    },
  });

  useEffect(() => {
    if (deviceStatus?.ipAddress) {
      const ip = deviceStatus.ipAddress.split('/')[0] || '198.18.0.1';
      setFormData(prev => ({
        ...prev,
        gateway: ip,
      }));
    }
  }, [deviceStatus]);

  const addMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      let cidr = data.cidr;
      if (!cidr.includes('/')) {
        cidr = `${cidr}/32`;
      }
      await api.post('/routes', { ...data, cidr });
    },
    onSuccess: () => {
      toast.success('Route added');
      queryClient.invalidateQueries({ queryKey: ['routes'] });
      setIsAddOpen(false);
      setFormData(prev => ({ ...prev, cidr: '' }));
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to add route');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (cidr: string) => {
      const encoded = encodeURIComponent(cidr);
      await api.delete(`/routes/${encoded}`);
    },
    onSuccess: () => {
      toast.success('Route deleted');
      queryClient.invalidateQueries({ queryKey: ['routes'] });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to delete route');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    addMutation.mutate(formData);
  };

  const handleCidrBlur = (value: string) => {
    if (value && !value.includes('/')) {
      setFormData(prev => ({
        ...prev,
        cidr: `${value}/32`,
      }));
    }
  };

  return (
    <>
      <div className="rounded-xl bg-white p-6 shadow-sm">
        <div className="flex items-center justify-between pb-4 mb-4">
          <div>
            <h3 className="font-semibold text-gray-900">Route Management</h3>
            <p className="text-sm text-gray-600 mt-1">Add and remove network routes</p>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => refetch()}
              className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md transition-colors"
              title="Refresh routes"
            >
              <RefreshCw className={`w-5 h-5 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
            <button
              type="button"
              onClick={() => setIsAddOpen(true)}
              className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md text-sm transition-colors"
            >
              <Plus className="w-4 h-4" />
              Add Route
            </button>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead className="bg-gray-50 text-gray-600 font-medium">
              <tr>
                <th className="px-4 py-3">Destination (CIDR)</th>
                <th className="px-4 py-3">Gateway</th>
                <th className="px-4 py-3">Device</th>
                <th className="px-4 py-3">Metric</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {isLoading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-gray-500">
                    Loading routes...
                  </td>
                </tr>
              ) : routes?.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-gray-500">
                    No routes found. Click "Add Route" to add a new route.
                  </td>
                </tr>
              ) : (
                routes?.map((route) => (
                  <tr key={route.cidr} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-mono text-gray-900">{route.cidr}</td>
                    <td className="px-4 py-3 font-mono text-gray-600">{route.gateway || '-'}</td>
                    <td className="px-4 py-3 text-gray-600">{route.device}</td>
                    <td className="px-4 py-3 text-gray-600">{route.metric}</td>
                    <td className="px-4 py-3 text-right">
                      <button
                        type="button"
                        onClick={() => deleteMutation.mutate(route.cidr)}
                        disabled={deleteMutation.isPending}
                        className="text-red-500 hover:text-red-700 hover:bg-red-50 p-2 rounded-md transition-colors"
                        title="Delete route"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Add Route Modal */}
      {isAddOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
          <div className="bg-white rounded-lg shadow-lg w-full max-w-md animate-in fade-in zoom-in-95">
            <div className="flex items-center justify-between p-6 border-b">
              <h2 className="text-lg font-semibold text-gray-900">Add New Route</h2>
              <button
                type="button"
                onClick={() => setIsAddOpen(false)}
                className="text-gray-400 hover:text-gray-600 transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Destination CIDR
                </label>
                <input
                  type="text"
                  name="cidr"
                  value={formData.cidr}
                  onChange={(e) => setFormData(prev => ({ ...prev, cidr: e.target.value }))}
                  onBlur={(e) => handleCidrBlur(e.target.value)}
                  placeholder="e.g. 192.168.1.0/24"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
                <p className="mt-1 text-xs text-gray-500">
                  Network CIDR notation. If no prefix is specified, /32 will be used.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Gateway
                </label>
                <input
                  type="text"
                  name="gateway"
                  value={formData.gateway}
                  readOnly
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm bg-gray-50 focus:outline-none text-gray-600"
                  title="Gateway is automatically filled with TUN device IP address"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Automatically filled with TUN device IP address (read-only)
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Metric
                </label>
                <input
                  type="number"
                  name="metric"
                  value={formData.metric}
                  onChange={(e) => setFormData(prev => ({ ...prev, metric: parseInt(e.target.value) || 100 }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  min="0"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsAddOpen(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={addMutation.isPending}
                  className="px-4 py-2 text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {addMutation.isPending ? 'Adding...' : 'Add Route'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
