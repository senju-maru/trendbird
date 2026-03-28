export interface ApiResponse<T> {
  data: T;
  success: boolean;
  error?: string;
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  total: number;
  page: number;
  limit: number;
}

export interface DashboardStats {
  detections: number;
  generations: number;
  lastCheckedAt: string | null;
}
