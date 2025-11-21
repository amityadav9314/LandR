import React from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, ActivityIndicator, RefreshControl } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { learningClient } from '../services/api';
import { useAuthStore } from '../store/authStore';
import { Flashcard } from '../../proto/learning';

// Define navigation types
type RootStackParamList = {
    Home: undefined;
    AddMaterial: undefined;
    Review: { flashcardId: string };
};

type HomeScreenNavigationProp = NativeStackNavigationProp<RootStackParamList, 'Home'>;

export const HomeScreen = () => {
    const navigation = useNavigation<HomeScreenNavigationProp>();
    const { user, logout } = useAuthStore();

    const { data, isLoading, error, refetch } = useQuery({
        queryKey: ['dueFlashcards'],
        queryFn: async () => {
            console.log('[HOME] Fetching due flashcards...');
            try {
                const response = await learningClient.getDueFlashcards({});
                console.log('[HOME] Got flashcards:', response.flashcards?.length || 0);
                return response.flashcards;
            } catch (err) {
                console.error('[HOME] Failed to fetch flashcards:', err);
                // Return empty array on error so UI still renders
                return [];
            }
        },
    });

    const renderItem = ({ item }: { item: Flashcard }) => (
        <TouchableOpacity
            style={styles.card}
            onPress={() => navigation.navigate('Review', { flashcardId: item.id })}
        >
            <Text style={styles.question}>{item.question}</Text>
            <Text style={styles.stage}>Stage: {item.stage}</Text>
        </TouchableOpacity>
    );

    return (
        <View style={styles.container}>
            <View style={styles.header}>
                <Text style={styles.title}>Welcome, {user?.name}</Text>
                <TouchableOpacity onPress={logout}>
                    <Text style={styles.logout}>Logout</Text>
                </TouchableOpacity>
            </View>

            <Text style={styles.sectionTitle}>Due for Review</Text>

            {error ? (
                <Text style={styles.error}>Failed to load flashcards</Text>
            ) : (
                <FlatList
                    data={data}
                    renderItem={renderItem}
                    keyExtractor={(item) => item.id}
                    contentContainerStyle={styles.list}
                    refreshControl={
                        <RefreshControl refreshing={isLoading} onRefresh={refetch} />
                    }
                    ListEmptyComponent={
                        <Text style={styles.empty}>No flashcards due! Good job.</Text>
                    }
                />
            )}

            <TouchableOpacity
                style={styles.fab}
                onPress={() => navigation.navigate('AddMaterial')}
            >
                <Text style={styles.fabText}>+</Text>
            </TouchableOpacity>
        </View>
    );
};

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#f5f5f5',
        paddingTop: 50,
        paddingHorizontal: 20,
    },
    center: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
    },
    header: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 20,
    },
    title: {
        fontSize: 20,
        fontWeight: 'bold',
        color: '#333',
    },
    logout: {
        color: '#d9534f',
        fontWeight: '600',
    },
    sectionTitle: {
        fontSize: 18,
        fontWeight: '600',
        marginBottom: 10,
        color: '#555',
    },
    list: {
        paddingBottom: 80,
    },
    card: {
        backgroundColor: '#fff',
        padding: 15,
        borderRadius: 10,
        marginBottom: 10,
        elevation: 2,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 1 },
        shadowOpacity: 0.2,
        shadowRadius: 1.41,
    },
    question: {
        fontSize: 16,
        color: '#333',
        marginBottom: 5,
    },
    stage: {
        fontSize: 12,
        color: '#888',
    },
    error: {
        color: 'red',
        textAlign: 'center',
        marginTop: 20,
    },
    empty: {
        textAlign: 'center',
        color: '#888',
        marginTop: 50,
        fontStyle: 'italic',
    },
    fab: {
        position: 'absolute',
        bottom: 30,
        right: 30,
        backgroundColor: '#4285F4',
        width: 56,
        height: 56,
        borderRadius: 28,
        justifyContent: 'center',
        alignItems: 'center',
        elevation: 5,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.3,
        shadowRadius: 3,
    },
    fabText: {
        color: '#fff',
        fontSize: 30,
        fontWeight: 'bold',
        marginTop: -2,
    },
});
