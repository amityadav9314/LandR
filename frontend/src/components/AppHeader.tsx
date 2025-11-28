import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAuthStore } from '../store/authStore';

export const AppHeader = () => {
    const navigation = useNavigation();
    const { logout } = useAuthStore();

    const handleHomePress = () => {
        // Navigate to home screen
        navigation.navigate('Home' as never);
    };

    return (
        <View style={styles.logoContainer}>
            <TouchableOpacity style={styles.homeIcon} onPress={handleHomePress}>
                <Text style={styles.homeIconText}>🏠</Text>
            </TouchableOpacity>
            <Text style={styles.logo}>LandR</Text>
            <TouchableOpacity onPress={logout} style={styles.logoutButton}>
                <Text style={styles.logout}>Logout</Text>
            </TouchableOpacity>
        </View>
    );
};

const styles = StyleSheet.create({
    logoContainer: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginTop: 5,
        marginBottom: 15,
        paddingHorizontal: 20,
        backgroundColor: '#f5f5f5',
    },
    homeIcon: {
        padding: 8,
    },
    homeIconText: {
        fontSize: 24,
    },
    logo: {
        fontSize: 28,
        fontWeight: 'bold',
        color: '#4285F4',
        letterSpacing: 1,
        position: 'absolute',
        left: '50%',
        transform: [{ translateX: -35 }],
    },
    logoutButton: {
        padding: 8,
    },
    logout: {
        color: '#d9534f',
        fontWeight: '600',
    },
});

