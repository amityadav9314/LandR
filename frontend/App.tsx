import React, { useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AppNavigator } from './src/navigation/AppNavigator';
import { StatusBar } from 'expo-status-bar';
import { NotificationService } from './src/services/notificationService';
import { useAuthStore } from './src/store/authStore';

const queryClient = new QueryClient();

function AppContent() {
  const { user } = useAuthStore();

  useEffect(() => {
    // Initialize notifications when user is logged in
    if (user) {
      NotificationService.initialize();
    }
  }, [user]);

  return (
    <>
      <AppNavigator />
      <StatusBar style="auto" />
    </>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppContent />
    </QueryClientProvider>
  );
}
