'use client';

import { useState, useMemo } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { C, up, dn, gradientBlue } from '@/lib/design-tokens';

export interface DateTimePickerProps {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  minDateTime?: string;
  error?: string;
}

const WEEKDAYS = ['月', '火', '水', '木', '金', '土', '日'] as const;

interface CalendarDay {
  date: number;
  month: number; // 0-11
  year: number;
  dateStr: string; // "YYYY-MM-DD"
  isCurrentMonth: boolean;
  isDisabled: boolean;
  isToday: boolean;
  isSelected: boolean;
}

/**
 * DateTimePicker — custom calendar + hour grid.
 * value format: "YYYY-MM-DDTHH:00"
 * minDateTime format: "YYYY-MM-DDTHH:00"
 */
export function DateTimePicker({
  label,
  value,
  onChange,
  minDateTime,
  error,
}: DateTimePickerProps) {
  const selectedDate = value ? value.slice(0, 10) : '';
  const selectedHour = value ? parseInt(value.slice(11, 13), 10) : -1;

  const minDate = minDateTime ? minDateTime.slice(0, 10) : '';
  const minHour = minDateTime ? parseInt(minDateTime.slice(11, 13), 10) : 0;

  const initialYear = value ? parseInt(value.slice(0, 4), 10) : new Date().getFullYear();
  const initialMonth = value ? parseInt(value.slice(5, 7), 10) - 1 : new Date().getMonth();

  const [viewYear, setViewYear] = useState(initialYear);
  const [viewMonth, setViewMonth] = useState(initialMonth);
  const [hovDay, setHovDay] = useState<string | null>(null);
  const [hovHour, setHovHour] = useState<number | null>(null);
  const [hovNav, setHovNav] = useState<'prev' | 'next' | null>(null);

  const today = useMemo(() => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  }, []);

  const calendarDays = useMemo<CalendarDay[]>(() => {
    const firstDay = new Date(viewYear, viewMonth, 1);
    // Monday=0 ... Sunday=6
    const startDow = (firstDay.getDay() + 6) % 7;
    const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate();

    const days: CalendarDay[] = [];

    // Previous month padding
    if (startDow > 0) {
      const prevMonthDays = new Date(viewYear, viewMonth, 0).getDate();
      const prevMonth = viewMonth === 0 ? 11 : viewMonth - 1;
      const prevYear = viewMonth === 0 ? viewYear - 1 : viewYear;
      for (let i = startDow - 1; i >= 0; i--) {
        const d = prevMonthDays - i;
        const dateStr = `${prevYear}-${String(prevMonth + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
        days.push({
          date: d,
          month: prevMonth,
          year: prevYear,
          dateStr,
          isCurrentMonth: false,
          isDisabled: minDate ? dateStr < minDate : false,
          isToday: dateStr === today,
          isSelected: dateStr === selectedDate,
        });
      }
    }

    // Current month
    for (let d = 1; d <= daysInMonth; d++) {
      const dateStr = `${viewYear}-${String(viewMonth + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
      days.push({
        date: d,
        month: viewMonth,
        year: viewYear,
        dateStr,
        isCurrentMonth: true,
        isDisabled: minDate ? dateStr < minDate : false,
        isToday: dateStr === today,
        isSelected: dateStr === selectedDate,
      });
    }

    // Next month padding
    const remaining = 7 - (days.length % 7);
    if (remaining < 7) {
      const nextMonth = viewMonth === 11 ? 0 : viewMonth + 1;
      const nextYear = viewMonth === 11 ? viewYear + 1 : viewYear;
      for (let d = 1; d <= remaining; d++) {
        const dateStr = `${nextYear}-${String(nextMonth + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
        days.push({
          date: d,
          month: nextMonth,
          year: nextYear,
          dateStr,
          isCurrentMonth: false,
          isDisabled: minDate ? dateStr < minDate : false,
          isToday: dateStr === today,
          isSelected: dateStr === selectedDate,
        });
      }
    }

    return days;
  }, [viewYear, viewMonth, minDate, today, selectedDate]);

  const canGoPrev = useMemo(() => {
    if (!minDate) return true;
    const minYear = parseInt(minDate.slice(0, 4), 10);
    const minMonth = parseInt(minDate.slice(5, 7), 10) - 1;
    return viewYear > minYear || (viewYear === minYear && viewMonth > minMonth);
  }, [viewYear, viewMonth, minDate]);

  const goToPrevMonth = () => {
    if (!canGoPrev) return;
    if (viewMonth === 0) {
      setViewYear(viewYear - 1);
      setViewMonth(11);
    } else {
      setViewMonth(viewMonth - 1);
    }
  };

  const goToNextMonth = () => {
    if (viewMonth === 11) {
      setViewYear(viewYear + 1);
      setViewMonth(0);
    } else {
      setViewMonth(viewMonth + 1);
    }
  };

  const handleDateChange = (dateStr: string) => {
    const startHour = dateStr === minDate ? minHour : 0;
    const hour = selectedHour >= startHour ? selectedHour : startHour;
    onChange(`${dateStr}T${String(hour).padStart(2, '0')}:00`);
  };

  const handleHourChange = (newHour: number) => {
    if (!selectedDate) return;
    onChange(`${selectedDate}T${String(newHour).padStart(2, '0')}:00`);
  };

  const isHourDisabled = (hour: number) => {
    if (!selectedDate) return true;
    if (selectedDate === minDate && hour < minHour) return true;
    return false;
  };

  const hasSelectedDate = selectedDate !== '';

  const navBtnStyle = (side: 'prev' | 'next', disabled?: boolean): React.CSSProperties => ({
    width: 32,
    height: 32,
    borderRadius: 8,
    border: 'none',
    background: C.bg,
    boxShadow: hovNav === side && !disabled ? up(3) : up(2),
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: disabled ? 'default' : 'pointer',
    opacity: disabled ? 0.3 : 1,
    transform: hovNav === side && !disabled ? 'translateY(-1px)' : 'none',
    transition: 'all 0.18s ease',
    color: C.text,
  });

  const dayCellStyle = (day: CalendarDay): React.CSSProperties => {
    const isHovered = hovDay === day.dateStr && !day.isDisabled;
    const base: React.CSSProperties = {
      width: 36,
      height: 36,
      borderRadius: 10,
      border: 'none',
      background: day.isSelected ? gradientBlue : C.bg,
      color: day.isSelected
        ? '#fff'
        : day.isDisabled
          ? C.textMuted
          : !day.isCurrentMonth
            ? C.textMuted
            : day.isToday
              ? C.blue
              : C.text,
      fontSize: 13,
      fontWeight: day.isToday || day.isSelected ? 600 : 400,
      cursor: day.isDisabled ? 'default' : 'pointer',
      opacity: day.isDisabled ? 0.3 : 1,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      transition: 'all 0.18s ease',
      fontFamily: 'inherit',
    };

    if (day.isSelected) {
      base.boxShadow = 'none';
      base.transform = 'none';
    } else if (isHovered) {
      base.boxShadow = up(2);
      base.transform = 'translateY(-1px)';
    } else if (day.isToday) {
      base.boxShadow = dn(1);
    } else {
      base.boxShadow = 'none';
    }

    return base;
  };

  const hourCellStyle = (hour: number): React.CSSProperties => {
    const isSelected = hasSelectedDate && selectedHour === hour;
    const disabled = isHourDisabled(hour);
    const isHovered = hovHour === hour && !disabled && hasSelectedDate;

    return {
      height: 36,
      borderRadius: 10,
      border: 'none',
      background: isSelected ? gradientBlue : C.bg,
      boxShadow: isSelected ? 'none' : isHovered ? up(3) : up(2),
      color: isSelected ? '#fff' : disabled ? C.textMuted : C.text,
      fontSize: 13,
      fontWeight: isSelected ? 600 : 400,
      cursor: disabled || !hasSelectedDate ? 'default' : 'pointer',
      opacity: disabled ? 0.3 : 1,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      transform: isHovered ? 'translateY(-1px)' : 'none',
      transition: 'all 0.18s ease',
      fontFamily: 'inherit',
    };
  };

  return (
    <div style={{ marginBottom: 18 }}>
      {label && (
        <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: C.textSub, marginBottom: 6 }}>
          {label}
        </label>
      )}

      <div
        style={{
          borderRadius: 16,
          background: C.bg,
          boxShadow: error ? `${dn(3)}, 0 0 0 3px ${C.red}` : dn(3),
          padding: 16,
          transition: 'box-shadow 0.22s ease',
        }}
      >
        {/* Month navigation */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
          <button
            type="button"
            onClick={goToPrevMonth}
            disabled={!canGoPrev}
            onMouseEnter={() => setHovNav('prev')}
            onMouseLeave={() => setHovNav(null)}
            style={navBtnStyle('prev', !canGoPrev)}
          >
            <ChevronLeft size={16} />
          </button>
          <span style={{ fontSize: 15, fontWeight: 600, color: C.text }}>
            {viewYear}年 {viewMonth + 1}月
          </span>
          <button
            type="button"
            onClick={goToNextMonth}
            onMouseEnter={() => setHovNav('next')}
            onMouseLeave={() => setHovNav(null)}
            style={navBtnStyle('next')}
          >
            <ChevronRight size={16} />
          </button>
        </div>

        {/* Weekday header */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 4, marginBottom: 4 }}>
          {WEEKDAYS.map((wd) => (
            <div
              key={wd}
              style={{
                textAlign: 'center',
                fontSize: 11,
                fontWeight: 500,
                color: C.textMuted,
                height: 28,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              {wd}
            </div>
          ))}
        </div>

        {/* Calendar grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 4 }}>
          {calendarDays.map((day) => (
            <button
              key={day.dateStr}
              type="button"
              disabled={day.isDisabled}
              onClick={() => !day.isDisabled && handleDateChange(day.dateStr)}
              onMouseEnter={() => setHovDay(day.dateStr)}
              onMouseLeave={() => setHovDay(null)}
              style={dayCellStyle(day)}
            >
              {day.date}
            </button>
          ))}
        </div>

        {/* Separator */}
        <div
          style={{
            height: 2,
            borderRadius: 1,
            background: C.bg,
            boxShadow: dn(1),
            margin: '16px 0',
          }}
        />

        {/* Time grid */}
        <div style={{ opacity: hasSelectedDate ? 1 : 0.35, pointerEvents: hasSelectedDate ? 'auto' : 'none', transition: 'opacity 0.18s ease' }}>
          <div style={{ fontSize: 12, fontWeight: 500, color: C.textSub, marginBottom: 8 }}>
            時刻を選択
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gap: 6 }}>
            {Array.from({ length: 24 }, (_, h) => (
              <button
                key={h}
                type="button"
                disabled={isHourDisabled(h)}
                onClick={() => !isHourDisabled(h) && handleHourChange(h)}
                onMouseEnter={() => setHovHour(h)}
                onMouseLeave={() => setHovHour(null)}
                style={hourCellStyle(h)}
              >
                {h}
              </button>
            ))}
          </div>
        </div>
      </div>

      {error && (
        <div style={{ fontSize: 11, color: C.red, marginTop: 4 }}>{error}</div>
      )}
    </div>
  );
}
