import React from 'react';
import { Outlet } from 'react-router-dom';

export default function StandaloneLayout() {
  return (
    <div className="bg-surface dark:bg-ds-dark-surface min-h-screen text-on-surface dark:text-ds-dark-on-surface transition-colors duration-500 font-body relative flex flex-col items-center justify-center">
      <Outlet />
    </div>
  );
}
