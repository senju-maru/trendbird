'use client';

import { Modal } from './Modal';
import { Button } from './Button';
import { C } from '@/lib/design-tokens';

export interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'default' | 'danger';
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = '確認',
  cancelLabel = 'キャンセル',
  variant = 'default',
}: ConfirmDialogProps) {
  return (
    <Modal isOpen={isOpen} onClose={onClose} size="sm">
      <h3 style={{
        margin: '0 0 8px', fontSize: 18, fontWeight: 600,
        color: variant === 'danger' ? C.red : C.text,
      }}>
        {title}
      </h3>
      {description && (
        <p style={{ fontSize: 13, color: C.textSub, lineHeight: 1.7, margin: '0 0 20px' }}>
          {description}
        </p>
      )}
      <div style={{ display: 'flex', gap: 10 }}>
        <Button variant="ghost" size="md" fullWidth onClick={onClose}>{cancelLabel}</Button>
        <Button
          variant={variant === 'danger' ? 'destructive' : 'filled'}
          size="md"
          fullWidth
          onClick={onConfirm}
        >
          {confirmLabel}
        </Button>
      </div>
    </Modal>
  );
}
