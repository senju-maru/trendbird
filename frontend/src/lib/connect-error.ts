import { ConnectError, Code } from '@connectrpc/connect';
import { TRACE_ID_HEADER } from './trace';

const GENERIC_MESSAGE = 'ただいま処理がうまくいきませんでした。しばらく時間をおいて再度お試しください。改善されない場合はお問い合わせください';

const codeMessages: Record<number, string> = {
  // ユーザー起因（そのまま表示）
  [Code.InvalidArgument]: '入力内容に誤りがあります',
  [Code.NotFound]: 'データが見つかりません',
  [Code.AlreadyExists]: '既に登録されています',
  [Code.PermissionDenied]: '権限がありません',
  [Code.Unauthenticated]: '認証が必要です。再度ログインしてください',
  [Code.ResourceExhausted]: '利用上限に達しました',
  [Code.FailedPrecondition]: '操作を実行できません',
  // サーバー起因（ユーザーフレンドリーなメッセージ）
  [Code.Internal]: GENERIC_MESSAGE,
  [Code.Unavailable]: '現在サービスに接続しづらい状態です。しばらく時間をおいて再度お試しください',
  [Code.DeadlineExceeded]: '処理に時間がかかりすぎました。しばらく時間をおいて再度お試しください',
};

export function connectErrorToMessage(err: unknown): string {
  if (err instanceof ConnectError) {
    const traceId = err.metadata.get(TRACE_ID_HEADER);
    if (traceId) {
      console.error(`[ConnectError] code=${err.code} traceId=${traceId}`, err);
    }

    return codeMessages[err.code] ?? GENERIC_MESSAGE;
  }
  if (err instanceof Error) {
    console.error('[UnexpectedError]', err);
    return GENERIC_MESSAGE;
  }
  return GENERIC_MESSAGE;
}
