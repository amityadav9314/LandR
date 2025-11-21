import React, { useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ActivityIndicator } from 'react-native';
import * as WebBrowser from 'expo-web-browser';
import * as Google from 'expo-auth-session/providers/google';
import { useAuthStore } from '../store/authStore';
import { authClient } from '../services/api';
import { GOOGLE_WEB_CLIENT_ID, GOOGLE_IOS_CLIENT_ID, GOOGLE_ANDROID_CLIENT_ID } from '../utils/config';

WebBrowser.maybeCompleteAuthSession();

export const LoginScreen = () => {
    const { login } = useAuthStore();
    const [request, response, promptAsync] = Google.useAuthRequest({
        clientId: GOOGLE_WEB_CLIENT_ID,
        iosClientId: GOOGLE_IOS_CLIENT_ID,
        androidClientId: GOOGLE_ANDROID_CLIENT_ID,
    });

    useEffect(() => {
        if (response?.type === 'success') {
            const { id_token } = response.params;
            if (id_token) {
                handleLogin(id_token);
            }
        }
    }, [response]);

    const handleLogin = async (idToken: string) => {
        try {
            const result = await authClient.login({ googleIdToken: idToken });
            if (result.user && result.sessionToken) {
                await login(result.user, result.sessionToken);
            }
        } catch (error) {
            console.error("Login failed", error);
            // Show error to user
        }
    };

    return (
        <View style={styles.container}>
            <Text style={styles.title}>Learn & Revise</Text>
            <Text style={styles.subtitle}>Master any topic with AI and Spaced Repetition</Text>

            <TouchableOpacity
                style={styles.button}
                onPress={() => promptAsync()}
                disabled={!request}
            >
                <Text style={styles.buttonText}>Sign in with Google</Text>
            </TouchableOpacity>
        </View>
    );
};

const styles = StyleSheet.create({
    container: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        backgroundColor: '#fff',
        padding: 20,
    },
    title: {
        fontSize: 32,
        fontWeight: 'bold',
        marginBottom: 10,
        color: '#333',
    },
    subtitle: {
        fontSize: 16,
        color: '#666',
        marginBottom: 40,
        textAlign: 'center',
    },
    button: {
        backgroundColor: '#4285F4',
        paddingVertical: 12,
        paddingHorizontal: 24,
        borderRadius: 8,
        elevation: 2,
    },
    buttonText: {
        color: '#fff',
        fontSize: 16,
        fontWeight: '600',
    },
});
