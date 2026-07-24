export type Notification = {
  id: string;
  type: string;
  title: string;
  message: string;
  resource_type: string;
  resource_id: string;
  application_id: string;
  read_at?: string;
  created_at: string;
};
