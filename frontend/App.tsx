import 'react-native-gesture-handler';
import { enableScreens } from 'react-native-screens';
enableScreens(false);
import React, { useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AppNavigator } from './src/navigation/AppNavigator';
import { StatusBar } from 'expo-status-bar';
import { NotificationService } from './src/services/notificationService';
import { useAuthStore } from './src/store/authStore';
import { LoginScreen } from './src/screens/LoginScreen';
import { HomeScreen } from './src/screens/HomeScreen';
import { AddMaterialScreen } from './src/screens/AddMaterialScreen';
import { MaterialDetailScreen } from './src/screens/MaterialDetailScreen';
import { ReviewScreen } from './src/screens/ReviewScreen';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { NavigationProvider, useNavigation } from './src/navigation/ManualRouter';

const queryClient = new QueryClient();

const ScreenRenderer = () => {
  const { currentScreen } = useNavigation();

  switch (currentScreen) {
    case 'Home': return <HomeScreen />;
    case 'AddMaterial': return <AddMaterialScreen />;
    case 'MaterialDetail': return <MaterialDetailScreen />;
    case 'Review': return <ReviewScreen />;
    default: return <HomeScreen />;
  }
};

function AppContent() {
  const { user, restoreSession, isLoading } = useAuthStore();

  useEffect(() => {
    restoreSession();
  }, []);

  useEffect(() => {
    // Initialize notifications when user is logged in
    if (user) {
      // NotificationService.initialize();
    }
  }, [user]);

  if (isLoading) {
    return null; // Or a splash screen
  }

  // ...

  return (
    <SafeAreaProvider>
      <GestureHandlerRootView style={{ flex: 1 }}>
        {user ? (
          <NavigationProvider>
            <ScreenRenderer />
          </NavigationProvider>
        ) : (
          <LoginScreen />
        )}
      </GestureHandlerRootView>
      {/* <StatusBar style="auto" /> */}
    </SafeAreaProvider>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppContent />
    </QueryClientProvider>
  );
}
