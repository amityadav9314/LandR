import React, { useState } from 'react';
import { View, Text, StyleSheet, TextInput, TouchableOpacity, ScrollView, ActivityIndicator, Alert } from 'react-native';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigation } from '../navigation/ManualRouter';
import { learningClient } from '../services/api';
import { AppHeader } from '../components/AppHeader';
import { useTheme, ThemeColors } from '../utils/theme';

export const AddMaterialScreen = () => {
    const navigation = useNavigation();
    const queryClient = useQueryClient();
    const { colors } = useTheme();
    const [content, setContent] = useState('');
    const [type, setType] = useState<'TEXT' | 'LINK'>('TEXT');

    const mutation = useMutation({
        mutationFn: async () => {
            return await learningClient.addMaterial({
                type,
                content,
            });
        },
        onSuccess: (data) => {
            console.log('[ADD_MATERIAL] Success! Created flashcards:', data.flashcardsCreated);
            Alert.alert('Success', `Created ${data.flashcardsCreated} flashcards for "${data.title}"!`);
            queryClient.invalidateQueries({ queryKey: ['dueMaterials'], refetchType: 'all' });
            queryClient.invalidateQueries({ queryKey: ['dueFlashcards'], refetchType: 'all' });
            queryClient.invalidateQueries({ queryKey: ['allTags'], refetchType: 'all' });
            queryClient.invalidateQueries({ queryKey: ['notificationStatus'], refetchType: 'all' });
            navigation.goBack();
        },
        onError: (error) => {
            console.error(error);
            Alert.alert('Error', 'Failed to add material. Please try again.');
        },
    });

    const handleSubmit = () => {
        if (!content.trim()) {
            Alert.alert('Validation', 'Please enter some content');
            return;
        }
        mutation.mutate();
    };

    const styles = createStyles(colors);

    return (
        <View style={styles.container}>
            <AppHeader />
            <View style={styles.contentContainer}>
                <Text style={styles.title}>Add New Material</Text>

                <View style={styles.typeContainer}>
                    <TouchableOpacity
                        style={[styles.typeButton, type === 'TEXT' && styles.activeType]}
                        onPress={() => setType('TEXT')}
                    >
                        <Text style={[styles.typeText, type === 'TEXT' && styles.activeTypeText]}>Text</Text>
                    </TouchableOpacity>
                    <TouchableOpacity
                        style={[styles.typeButton, type === 'LINK' && styles.activeType]}
                        onPress={() => setType('LINK')}
                    >
                        <Text style={[styles.typeText, type === 'LINK' && styles.activeTypeText]}>Link</Text>
                    </TouchableOpacity>
                </View>

                <TextInput
                    style={styles.input}
                    placeholder={type === 'TEXT' ? "Paste your text here..." : "Enter URL here..."}
                    placeholderTextColor={colors.textPlaceholder}
                    multiline={type === 'TEXT'}
                    numberOfLines={type === 'TEXT' ? 10 : 1}
                    value={content}
                    onChangeText={setContent}
                    textAlignVertical="top"
                />

                <TouchableOpacity
                    style={styles.submitButton}
                    onPress={handleSubmit}
                    disabled={mutation.isPending}
                >
                    {mutation.isPending ? (
                        <ActivityIndicator color={colors.textInverse} />
                    ) : (
                        <Text style={styles.submitText}>Generate Flashcards</Text>
                    )}
                </TouchableOpacity>
            </View>
        </View>
    );
};

const createStyles = (colors: ThemeColors) => StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: colors.background,
    },
    contentContainer: {
        flex: 1,
        padding: 20,
        backgroundColor: colors.card,
    },
    title: {
        fontSize: 24,
        fontWeight: 'bold',
        marginBottom: 20,
        color: colors.textPrimary,
    },
    typeContainer: {
        flexDirection: 'row',
        marginBottom: 20,
        backgroundColor: colors.cardAlt,
        borderRadius: 8,
        padding: 4,
    },
    typeButton: {
        flex: 1,
        paddingVertical: 10,
        alignItems: 'center',
        borderRadius: 6,
    },
    activeType: {
        backgroundColor: colors.card,
        elevation: 2,
    },
    typeText: {
        fontWeight: '600',
        color: colors.textSecondary,
    },
    activeTypeText: {
        color: colors.primary,
    },
    input: {
        borderWidth: 1,
        borderColor: colors.inputBorder,
        borderRadius: 8,
        padding: 15,
        fontSize: 16,
        marginBottom: 20,
        minHeight: 100,
        backgroundColor: colors.input,
        color: colors.textPrimary,
    },
    submitButton: {
        backgroundColor: colors.primary,
        paddingVertical: 15,
        borderRadius: 8,
        alignItems: 'center',
    },
    submitText: {
        color: colors.textInverse,
        fontSize: 16,
        fontWeight: 'bold',
    },
});
