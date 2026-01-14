export interface DeviceStatus {
  name: string;
  exists: boolean;
  status: 'up' | 'down' | 'unknown';
  ipAddress?: string;
  mtu?: number;
}

export interface Route {
  cidr: string;
  gateway: string;
  device: string;
  metric: number;
}

export interface ServiceStatus {
  running: boolean;
  pid?: number;
  uptime?: number;
  connections?: number;
  memoryUsage?: number;
  cpuUsage?: number;
  traffic?: TrafficStats;
}

export interface TrafficStats {
  uploadBytes: number;
  downloadBytes: number;
  uploadSpeed: number;
  downloadSpeed: number;
}

export interface ProxyConfig {
  type: 'socks5' | 'socks4' | 'http' | 'https';
  address: string;
  username?: string;
  password?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}

export interface AuthResponse {
  token: string;
  expiresIn: number;
}
