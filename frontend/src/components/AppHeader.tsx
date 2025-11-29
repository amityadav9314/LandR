import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert } from 'react-native';
// @ts-ignore
import { Ionicons } from '@expo/vector-icons';
// import { useNavigation } from '@react-navigation/native';
import { useNavigation } from '../navigation/ManualRouter';
import { useAuthStore } from '../store/authStore';

export const AppHeader = () => {
    const { navigate, goBack, canGoBack } = useNavigation();
    const { logout } = useAuthStore();

    const handleHomePress = () => {
        navigate('Home');
    };

    return (
        <View style={styles.headerContainer}>
            <View style={styles.leftContainer}>
                {canGoBack && (
                    <TouchableOpacity onPress={goBack} style={styles.iconButton}>
                        <Ionicons name="arrow-back" size={24} color="#333" />
                    </TouchableOpacity>
                )}
                <TouchableOpacity onPress={handleHomePress} style={styles.iconButton}>
                    <Text style={{ fontSize: 24 }}>🏠</Text>
                </TouchableOpacity>
            </View>

            <TouchableOpacity onPress={handleHomePress} style={styles.centerContainer}>
                <Text style={styles.logo}>LandR</Text>
            </TouchableOpacity>

            <View style={styles.rightContainer}>
                <TouchableOpacity onPress={logout} style={styles.logoutButton}>
                    <Text style={styles.logout}>Logout</Text>
                </TouchableOpacity>
            </View>
        </View>
    );
};

const styles = StyleSheet.create({
    headerContainer: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        paddingTop: 10,
        paddingBottom: 15,
        paddingHorizontal: 15,
        backgroundColor: '#f5f5f5',
        borderBottomWidth: 1,
        borderBottomColor: '#e0e0e0',
    },
    leftContainer: {
        flexDirection: 'row',
        alignItems: 'center',
        flex: 1,
    },
    centerContainer: {
        flex: 2,
        alignItems: 'center',
    },
    rightContainer: {
        flex: 1,
        alignItems: 'flex-end',
    },
    iconButton: {
        padding: 8,
        marginRight: 5,
    },
    logo: {
        fontSize: 24,
        fontWeight: 'bold',
        color: '#4285F4',
        letterSpacing: 1,
    },
    logoutButton: {
        padding: 8,
    },
    logout: {
        color: '#d9534f',
        fontWeight: '600',
        fontSize: 14,
    },
});

