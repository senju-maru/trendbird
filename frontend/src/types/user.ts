export interface User {
  id: string;
  name: string;           // Xから取得、読み取り専用
  email: string;          // 通知用、編集可能
  image: string;          // Xアバター、読み取り専用
  twitterHandle: string;  // Xハンドル、読み取り専用
  createdAt: string;
}
