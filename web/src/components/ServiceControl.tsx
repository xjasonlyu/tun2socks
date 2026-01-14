import { ArrowUp, ArrowDown } from 'lucide-react';
import type { ServiceStatus } from '@/api/types';

interface ServiceControlProps {
  status: ServiceStatus | undefined;
  isLoading: boolean;
}

export function ServiceControl({ status, isLoading }: ServiceControlProps) {
  if (isLoading) {
    return <div className="animate-pulse h-48 bg-gray-800 rounded-xl" />;
  }

  const isRunning = status?.running;

  return (
    <div className="rounded-xl bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between pb-4 mb-4">
        <div>
          <h3 className="font-semibold text-gray-900">Service Status</h3>
          <p className="text-sm text-gray-600 mt-1">tun2socks Engine Status</p>
        </div>
        <div className={`w-3 h-3 rounded-full ${isRunning ? 'bg-green-600' : 'bg-red-500'}`} />
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-gray-600">Status</span>
          <span className={`text-sm font-medium ${isRunning ? 'text-green-600' : 'text-red-500'}`}>
            {isRunning ? 'Running' : 'Stopped'}
          </span>
        </div>

        {isRunning && (
          <>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">PID</span>
              <span className="text-sm font-mono text-gray-900">{status?.pid}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600">Uptime</span>
              <span className="text-sm font-mono text-gray-900">{formatUptime(status?.uptime || 0)}</span>
            </div>

            <div className="pt-3">
              <h4 className="text-sm font-medium text-gray-900 mb-3">Traffic Statistics</h4>
              <div className="grid grid-cols-2 gap-3">
                <div className="bg-gray-50 p-3 rounded-lg">
                  <div className="flex items-center gap-2 text-sm text-gray-600 mb-1">
                    <ArrowUp className="w-4 h-4 text-green-600" />
                    <span>Upload</span>
                  </div>
                  <div className="text-lg font-mono font-semibold text-gray-900">
                    {formatBytes(status?.traffic?.uploadSpeed || 0)}/s
                  </div>
                </div>

                <div className="bg-gray-50 p-3 rounded-lg">
                  <div className="flex items-center gap-2 text-sm text-gray-600 mb-1">
                    <ArrowDown className="w-4 h-4 text-blue-600" />
                    <span>Download</span>
                  </div>
                  <div className="text-lg font-mono font-semibold text-gray-900">
                    {formatBytes(status?.traffic?.downloadSpeed || 0)}/s
                  </div>
                </div>
              </div>
            </div>
          </>
        )}

        {!isRunning && (
          <div className="pt-3">
            <p className="text-sm text-gray-600 italic">
              Service is currently stopped. Start the tun2socks engine to view statistics.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${h}h ${m}m ${s}s`;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
}
