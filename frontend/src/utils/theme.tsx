import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { Platform, useColorScheme } from 'react-native';

// Color definitions
export const lightColors = {
    // Backgrounds
    background: '#f5f5f5',
    card: '#ffffff',
    cardAlt: '#fafafa',
    input: '#fafafa',
    inputBorder: '#dddddd',

    // Text
    textPrimary: '#333333',
    textSecondary: '#666666',
    textPlaceholder: '#999999',
    textInverse: '#ffffff',

    // Brand colors
    primary: '#4285F4',
    primaryLight: '#a0cfff',
    success: '#34A853',
    error: '#d9534f',
    errorBg: '#ffebee',
    warning: '#ff4444',

    // Borders & Dividers
    border: '#dddddd',
    divider: '#eeeeee',

    // Header
    headerBg: '#f5f5f5',
    headerBorder: '#e0e0e0',

    // Tag chips
    tagBg: '#e0e0e0',
    tagText: '#555555',
    tagActiveBg: '#4285F4',
    tagActiveText: '#ffffff',

    // Badges
    badgeBg: '#4285F4',
    badgeText: '#ffffff',
    notificationBadgeBg: '#ff4444',

    // Pagination
    paginationDisabledBg: '#cccccc',
    paginationDisabledText: '#999999',

    // Filter chips
    filterChipBg: '#ffffff',
    filterChipBorder: '#dddddd',
};

export const darkColors = {
    // Backgrounds
    background: '#121212',
    card: '#1e1e1e',
    cardAlt: '#252525',
    input: '#2a2a2a',
    inputBorder: '#404040',

    // Text
    textPrimary: '#e0e0e0',
    textSecondary: '#a0a0a0',
    textPlaceholder: '#666666',
    textInverse: '#121212',

    // Brand colors
    primary: '#5a9cff',
    primaryLight: '#3d6ba8',
    success: '#4ade80',
    error: '#f87171',
    errorBg: '#3d1f1f',
    warning: '#ff6b6b',

    // Borders & Dividers
    border: '#333333',
    divider: '#2a2a2a',

    // Header
    headerBg: '#1a1a1a',
    headerBorder: '#333333',

    // Tag chips
    tagBg: '#333333',
    tagText: '#b0b0b0',
    tagActiveBg: '#5a9cff',
    tagActiveText: '#121212',

    // Badges
    badgeBg: '#5a9cff',
    badgeText: '#121212',
    notificationBadgeBg: '#ff6b6b',

    // Pagination
    paginationDisabledBg: '#333333',
    paginationDisabledText: '#666666',

    // Filter chips
    filterChipBg: '#2a2a2a',
    filterChipBorder: '#404040',
};

export type ThemeColors = typeof lightColors;

interface ThemeContextType {
    isDark: boolean;
    colors: ThemeColors;
    toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

// Storage helper for theme preference
const getStoredTheme = async (): Promise<boolean | null> => {
    try {
        if (Platform.OS === 'web') {
            const stored = localStorage.getItem('theme_dark');
            return stored ? stored === 'true' : null;
        } else {
            const SecureStore = await import('expo-secure-store');
            const stored = await SecureStore.getItemAsync('theme_dark');
            return stored ? stored === 'true' : null;
        }
    } catch {
        return null;
    }
};

const storeTheme = async (isDark: boolean): Promise<void> => {
    try {
        if (Platform.OS === 'web') {
            localStorage.setItem('theme_dark', isDark.toString());
        } else {
            const SecureStore = await import('expo-secure-store');
            await SecureStore.setItemAsync('theme_dark', isDark.toString());
        }
    } catch (error) {
        console.warn('[Theme] Failed to store theme preference:', error);
    }
};

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
    const systemColorScheme = useColorScheme();
    const [isDark, setIsDark] = useState(systemColorScheme === 'dark');
    const [isLoaded, setIsLoaded] = useState(false);

    // Load stored preference on mount
    useEffect(() => {
        const loadTheme = async () => {
            const storedTheme = await getStoredTheme();
            if (storedTheme !== null) {
                setIsDark(storedTheme);
            }
            setIsLoaded(true);
        };
        loadTheme();
    }, []);

    const toggleTheme = () => {
        const newValue = !isDark;
        setIsDark(newValue);
        storeTheme(newValue);
    };

    const colors = isDark ? darkColors : lightColors;

    // Don't render until theme is loaded to prevent flash
    if (!isLoaded) {
        return null;
    }

    return (
        <ThemeContext.Provider value={{ isDark, colors, toggleTheme }}>
            {children}
        </ThemeContext.Provider>
    );
};

export const useTheme = (): ThemeContextType => {
    const context = useContext(ThemeContext);
    if (!context) {
        throw new Error('useTheme must be used within a ThemeProvider');
    }
    return context;
};
