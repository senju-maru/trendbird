import { sendGAEvent } from '@next/third-parties/google';

const isEnabled = () => !!process.env.NEXT_PUBLIC_GA_ID;

// ── ユーザープロパティ ──
export function setUserProperties(userId: string) {
  if (!isEnabled() || typeof window === 'undefined') return;
  window.gtag?.('set', { user_id: userId });
}

// ── 認証 ──
export function trackLogin(method: string, isNewUser: boolean = false) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'login', { method, is_new_user: isNewUser });
}

export function trackSignup(method: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'sign_up', { method });
}

export function trackLogout() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'logout', {});
}

// ── トピック ──
export function trackTopicCreate(genre: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'topic_create', { genre });
}

export function trackTopicDelete() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'topic_delete', {});
}

export function trackGenreAdd(genre: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'genre_add', { genre });
}

// ── AI生成 ──
export function trackAiGenerate(topicId: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'ai_generate', { topic_id: topicId });
}

// ── 投稿 ──
export function trackPostPublish() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'post_publish', {});
}

export function trackPostSchedule() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'post_schedule', {});
}

export function trackDraftCreate() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'draft_create', {});
}

// ── LP ──
export function trackCtaClick(location: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'cta_click', { location, transport_type: 'beacon' });
}

// ── ログインページ ──
export function trackXLoginClick() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'x_login_click', {});
}

// ── AI結果アクション ──
export function trackAiResultCopy(source: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'ai_result_copy', { source });
}

export function trackAiResultSaveDraft() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'ai_result_save_draft', {});
}

// ── チュートリアル ──
export function trackTutorialStart(triggerSource: 'is_new_user' | 'session_pending') {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_start', { trigger_source: triggerSource });
}

export function trackTutorialComplete(totalSeconds: number) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_complete', { total_seconds: totalSeconds });
}

export function trackTutorialPhaseEnter(phase: string, fromPhase?: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_phase', { phase, from_phase: fromPhase });
}

export function trackTutorialGateCheck(result: 'started' | 'blocked_completed' | 'blocked_no_trigger') {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_gate_check', { result });
}

export function trackTutorialElementTimeout(phase: string, selector: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_element_timeout', { phase, selector, timeout_ms: 10000 });
}

export function trackTutorialDriverError(phase: string, error: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_driver_error', { phase, error });
}

export function trackTutorialAbandon(phase: string, secondsInPhase: number) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_abandon', { phase, seconds_in_phase: secondsInPhase });
}

export function trackTutorialBlockedStale(currentUserId: string) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'tutorial_blocked_stale', { current_user_id: currentUserId });
}

// ── 通知 ──
export function trackNotificationToggle(type: string, enabled: boolean) {
  if (!isEnabled()) return;
  sendGAEvent('event', 'notification_toggle', { type, enabled });
}

// ── Auto-DM ──
export function trackAutoDmRuleCreate() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'auto_dm_rule_create', {});
}

export function trackAutoDmRuleDelete() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'auto_dm_rule_delete', {});
}

// ── Auto-Reply ──
export function trackAutoReplyRuleCreate() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'auto_reply_rule_create', {});
}

export function trackAutoReplyRuleDelete() {
  if (!isEnabled()) return;
  sendGAEvent('event', 'auto_reply_rule_delete', {});
}
