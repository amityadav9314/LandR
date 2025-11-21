import React, { useState } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ActivityIndicator, Alert } from 'react-native';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRoute, useNavigation, RouteProp } from '@react-navigation/native';
import { learningClient } from '../services/api';
import { Flashcard } from '../../proto/learning';

type RootStackParamList = {
    Review: { flashcardId: string };
};

type ReviewScreenRouteProp = RouteProp<RootStackParamList, 'Review'>;

export const ReviewScreen = () => {
    const route = useRoute<ReviewScreenRouteProp>();
    const navigation = useNavigation();
    const queryClient = useQueryClient();
    const { flashcardId } = route.params;
    const [showAnswer, setShowAnswer] = useState(false);

    // Fetch the specific flashcard (or find it in cache)
    // Since we don't have getFlashcardById, we rely on the list or pass data.
    // But for correctness, we should probably pass the data or fetch it.
    // Given the MVP, let's assume we find it in the 'dueFlashcards' query cache.
    const flashcards = queryClient.getQueryData<Flashcard[]>(['dueFlashcards']);
    const flashcard = flashcards?.find(f => f.id === flashcardId);

    const mutation = useMutation({
        mutationFn: async () => {
            return await learningClient.completeReview({ flashcardId });
        },
        onSuccess: () => {
            Alert.alert('Great job!', 'Review completed.');
            queryClient.invalidateQueries({ queryKey: ['dueFlashcards'] });
            navigation.goBack();
        },
        onError: (error) => {
            console.error(error);
            Alert.alert('Error', 'Failed to complete review.');
        },
    });

    if (!flashcard) {
        return (
            <View style={styles.center}>
                <Text>Flashcard not found</Text>
            </View>
        );
    }

    return (
        <View style={styles.container}>
            <View style={styles.card}>
                <Text style={styles.label}>Question</Text>
                <Text style={styles.text}>{flashcard.question}</Text>

                {showAnswer && (
                    <View style={styles.answerContainer}>
                        <View style={styles.divider} />
                        <Text style={styles.label}>Answer</Text>
                        <Text style={styles.text}>{flashcard.answer}</Text>
                    </View>
                )}
            </View>

            {!showAnswer ? (
                <TouchableOpacity
                    style={styles.button}
                    onPress={() => setShowAnswer(true)}
                >
                    <Text style={styles.buttonText}>Show Answer</Text>
                </TouchableOpacity>
            ) : (
                <TouchableOpacity
                    style={[styles.button, styles.completeButton]}
                    onPress={() => mutation.mutate()}
                    disabled={mutation.isPending}
                >
                    {mutation.isPending ? (
                        <ActivityIndicator color="#fff" />
                    ) : (
                        <Text style={styles.buttonText}>Complete Review</Text>
                    )}
                </TouchableOpacity>
            )}
        </View>
    );
};

const styles = StyleSheet.create({
    container: {
        flex: 1,
        padding: 20,
        backgroundColor: '#f5f5f5',
        justifyContent: 'center',
    },
    center: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
    },
    card: {
        backgroundColor: '#fff',
        padding: 30,
        borderRadius: 15,
        elevation: 4,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 4,
        marginBottom: 30,
        minHeight: 200,
        justifyContent: 'center',
    },
    label: {
        fontSize: 14,
        color: '#888',
        marginBottom: 10,
        textTransform: 'uppercase',
        fontWeight: 'bold',
    },
    text: {
        fontSize: 22,
        color: '#333',
        textAlign: 'center',
        lineHeight: 30,
    },
    answerContainer: {
        marginTop: 20,
    },
    divider: {
        height: 1,
        backgroundColor: '#eee',
        marginVertical: 20,
    },
    button: {
        backgroundColor: '#4285F4',
        paddingVertical: 15,
        borderRadius: 30,
        alignItems: 'center',
        elevation: 2,
    },
    completeButton: {
        backgroundColor: '#34A853',
    },
    buttonText: {
        color: '#fff',
        fontSize: 18,
        fontWeight: 'bold',
    },
});
