import { useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { LogOut } from 'lucide-react';
import { api } from '@/api/client';
import type { ServiceStatus, DeviceStatus, ApiResponse } from '@/api/types';
import { ServiceControl } from '@/components/ServiceControl';
import { ProxySettings } from '@/components/ProxySettings';
import { RouteList } from '@/components/RouteList';

export function Dashboard() {
  const queryClient = useQueryClient();

  const { data: serviceStatus, isLoading: isServiceLoading } = useQuery({
    queryKey: ['service'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<ServiceStatus>>('/service');
      return res.data.data;
    },
  });

  useEffect(() => {
    const token = localStorage.getItem('auth-token');
    if (!token) return;

    const eventSource = new EventSource(`/api/v1/service/events?token=${token}`);

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        queryClient.setQueryData(['service'], data);
      } catch (e) {
        console.error('SSE parse error', e);
      }
    };

    return () => {
      eventSource.close();
    };
  }, [queryClient]);

  const handleLogout = () => {
    localStorage.removeItem('auth-token');
    window.location.href = '/login';
  };

  return (
    <div className="min-h-screen bg-white">
      {/* Top Navigation Bar */}
      <header className="border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-xl font-bold text-gray-900">tun2socks</h1>
              <p className="text-sm text-gray-600">Web Interface</p>
            </div>
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors px-3 py-2 rounded-md hover:bg-gray-100"
            >
              <LogOut className="w-4 h-4" />
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="space-y-6">
          {/* First Row: 3 Cards */}
          <div className="grid gap-6 md:grid-cols-3">
            <ServiceControl status={serviceStatus} isLoading={isServiceLoading} />
            <DeviceStatus queryKey={['device']} isLoading={false} />
            <ProxySettings />
          </div>

          {/* Second Row: Route Management */}
          <RouteList />
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-200 bg-white mt-12">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between text-sm text-gray-600">
            <span>© 2026 tun2socks Web Interface</span>
            <span>Built with React & TypeScript</span>
          </div>
        </div>
      </footer>
    </div>
  );
}

function DeviceStatus({ queryKey, isLoading }: { queryKey: any[]; isLoading: boolean }) {
  const { data: deviceStatus } = useQuery({
    queryKey,
    queryFn: async () => {
      const res = await api.get<ApiResponse<DeviceStatus>>('/device');
      return res.data.data;
    },
    refetchInterval: 5000,
  });

  if (isLoading) {
    return <div className="animate-pulse h-48 bg-gray-800 rounded-xl" />;
  }

  const isUp = deviceStatus?.status === 'up';

  return (
    <div className="rounded-xl bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between pb-4 mb-4">
        <div>
          <h3 className="font-semibold text-gray-900">TUN Device Status</h3>
          <p className="text-sm text-gray-600 mt-1">Network Interface Information</p>
        </div>
        <div className={`w-3 h-3 rounded-full ${isUp ? 'bg-green-600' : 'bg-red-500'}`} />
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-gray-600">Status</span>
          <span className={`text-sm font-medium ${isUp ? 'text-green-600' : 'text-red-500'}`}>
            {deviceStatus?.exists ? (isUp ? 'UP' : 'DOWN') : 'Not Created'}
          </span>
        </div>

        {deviceStatus?.exists && (
          <>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Device Name</span>
              <span className="text-sm font-mono text-gray-900">{deviceStatus?.name}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">IP Address</span>
              <span className="text-sm font-mono text-gray-900">{deviceStatus?.ipAddress || '-'}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">MTU</span>
              <span className="text-sm font-mono text-gray-900">{deviceStatus?.mtu || '-'}</span>
            </div>
          </>
        )}

        {!deviceStatus?.exists && (
          <div className="pt-3 border-t border-gray-200">
            <p className="text-sm text-gray-600 italic">
              TUN device has not been created yet. Start tun2socks to initialize the device.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
