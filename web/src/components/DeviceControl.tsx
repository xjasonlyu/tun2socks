import { Network } from 'lucide-react';
import type { DeviceStatus } from '@/api/types';

interface DeviceControlProps {
  status: DeviceStatus | undefined;
  isLoading: boolean;
}

export function DeviceControl({ status, isLoading }: DeviceControlProps) {
  if (isLoading) {
    return <div className="animate-pulse h-48 bg-gray-800 rounded-xl" />;
  }

  const isUp = status?.status === 'up';

  return (
    <div className="rounded-xl bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between pb-4 mb-4">
        <div>
          <h3 className="font-semibold text-gray-900">TUN Device Status</h3>
          <p className="text-sm text-gray-600 mt-1">Network Interface Information</p>
        </div>
        <Network className={`w-5 h-5 ${isUp ? 'text-green-600' : 'text-red-500'}`} />
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-gray-600">Status</span>
          <span className={`text-sm font-medium ${isUp ? 'text-green-600' : 'text-red-500'}`}>
            {status?.exists ? (isUp ? 'UP' : 'DOWN') : 'Not Created'}
          </span>
        </div>

        {status?.exists && (
          <>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Device Name</span>
              <span className="text-sm font-mono text-gray-900">{status?.name}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">IP Address</span>
              <span className="text-sm font-mono text-gray-900">{status?.ipAddress || '-'}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">MTU</span>
              <span className="text-sm font-mono text-gray-900">{status?.mtu || '-'}</span>
            </div>
          </>
        )}

        {!status?.exists && (
          <div className="pt-3">
            <p className="text-sm text-gray-600 italic">
              TUN device has not been created yet. Start tun2socks to initialize the device.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
