import {
  Monitor, Briefcase, Megaphone, TrendingUp, Palette,
  Coffee, Compass, Heart, BookOpen, Film,
  Bot, Video, Target, Code, Terminal,
} from 'lucide-react';

export const GENRE_ICONS: Record<string, React.ComponentType<{ size?: number; strokeWidth?: number }>> = {
  technology: Monitor,
  business: Briefcase,
  marketing: Megaphone,
  finance: TrendingUp,
  creative: Palette,
  lifestyle: Coffee,
  career: Compass,
  'health-beauty': Heart,
  education: BookOpen,
  entertainment: Film,
  // 旧スラグ（migration 000008）
  'ai-developer': Bot,
  'video-creator': Video,
  marketer: Target,
  'indie-creator': Code,
  engineer: Terminal,
};
