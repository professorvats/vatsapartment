'use client';

import * as React from 'react';
import { format } from 'date-fns';
import { Calendar as CalendarIcon } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Calendar } from '@/components/ui/calendar';
import { Popover } from '@/components/ui/popover';

interface DatePickerProps {
  value?: Date;
  onChange?: (date: Date | undefined) => void;
  placeholder?: string;
  className?: string;
}

export function DatePicker({
  value,
  onChange,
  placeholder = 'Pick a date',
  className,
}: DatePickerProps) {
  const [isOpen, setIsOpen] = React.useState(false);

  return (
    <div className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'w-full flex items-center justify-between px-4 py-3 md:px-4 md:py-4 text-base md:text-lg font-body-md text-primary-container bg-surface-container-low border border-outline-variant rounded-lg hover:border-outline transition-colors focus:outline-none focus:ring-2 focus:ring-primary-container/50 text-left',
          !value && 'text-outline'
        )}
      >
        <span className={cn(!value && 'text-outline')}>
          {value ? format(value, 'PPP') : placeholder}
        </span>
        <CalendarIcon className="w-5 h-5 text-on-surface-variant pointer-events-none" />
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40 bg-black/20"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute z-50 mt-1 w-full">
            <Popover>
              <Calendar
                mode="single"
                selected={value}
                onSelect={(date) => {
                  onChange?.(date);
                  setIsOpen(false);
                }}
                initialFocus
                disabled={(date) =>
                  date < new Date(new Date().setHours(0, 0, 0, 0))
                }
              />
            </Popover>
          </div>
        </>
      )}
    </div>
  );
}
