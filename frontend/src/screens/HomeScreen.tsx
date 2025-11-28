import React, { useState, useMemo } from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, RefreshControl, TextInput, ScrollView } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { learningClient } from '../services/api';
import { useAuthStore } from '../store/authStore';
import { MaterialSummary } from '../../proto/backend/proto/learning/learning';

// Define navigation types
type RootStackParamList = {
    Home: undefined;
    AddMaterial: undefined;
    MaterialDetail: { materialId: string; title: string };
};

type HomeScreenNavigationProp = NativeStackNavigationProp<RootStackParamList, 'Home'>;

export const HomeScreen = () => {
    const navigation = useNavigation<HomeScreenNavigationProp>();
    const { user, logout } = useAuthStore();
    
    // Filter states
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedTags, setSelectedTags] = useState<string[]>([]);

    const { data, isLoading, error, refetch } = useQuery({
        queryKey: ['dueMaterials'],
        queryFn: async () => {
            console.log('[HOME] Fetching due materials...');
            try {
                const response = await learningClient.getDueMaterials({});
                console.log('[HOME] Got materials:', response.materials?.length || 0);
                return response.materials;
            } catch (err) {
                console.error('[HOME] Failed to fetch materials:', err);
                return [];
            }
        },
    });

    // Fetch all tags from backend (works with pagination!)
    const { data: tagsData } = useQuery({
        queryKey: ['allTags'],
        queryFn: async () => {
            console.log('[HOME] Fetching all tags...');
            try {
                const response = await learningClient.getAllTags({});
                console.log('[HOME] Got tags:', response.tags?.length || 0);
                return response.tags || [];
            } catch (err) {
                console.error('[HOME] Failed to fetch tags:', err);
                return [];
            }
        },
    });

    const allTags = tagsData || [];

    // Filter materials based on search and selected tags
    const filteredMaterials = useMemo(() => {
        if (!data) return [];
        
        return data.filter(material => {
            // Filter by search query (title contains)
            const matchesSearch = searchQuery.trim() === '' || 
                material.title.toLowerCase().includes(searchQuery.toLowerCase());
            
            // Filter by selected tags (material must have ALL selected tags)
            const matchesTags = selectedTags.length === 0 || 
                selectedTags.every(tag => material.tags.includes(tag));
            
            return matchesSearch && matchesTags;
        });
    }, [data, searchQuery, selectedTags]);

    const toggleTag = (tag: string) => {
        setSelectedTags(prev => 
            prev.includes(tag) 
                ? prev.filter(t => t !== tag)
                : [...prev, tag]
        );
    };

    const clearFilters = () => {
        setSearchQuery('');
        setSelectedTags([]);
    };

    const renderItem = ({ item }: { item: MaterialSummary }) => (
        <TouchableOpacity
            style={styles.card}
            onPress={() => navigation.navigate('MaterialDetail', { materialId: item.id, title: item.title })}
        >
            <View style={styles.cardHeader}>
                <Text style={styles.materialTitle}>{item.title}</Text>
                <View style={styles.badge}>
                    <Text style={styles.badgeText}>{item.dueCount}</Text>
                </View>
            </View>

            <View style={styles.tagsContainer}>
                {item.tags.map((tag, index) => (
                    <View key={index} style={styles.tagBadge}>
                        <Text style={styles.tagText}>{tag}</Text>
                    </View>
                ))}
            </View>
        </TouchableOpacity>
    );

    const hasActiveFilters = searchQuery.trim() !== '' || selectedTags.length > 0;

    return (
        <View style={styles.container}>
            <View style={styles.header}>
                <Text style={styles.title}>Welcome, {user?.name}</Text>
                <TouchableOpacity onPress={logout}>
                    <Text style={styles.logout}>Logout</Text>
                </TouchableOpacity>
            </View>

            <Text style={styles.mainTitle}>Due for Review</Text>

            {/* Search Bar */}
            <View style={styles.searchContainer}>
                <TextInput
                    style={styles.searchInput}
                    placeholder="Search by title..."
                    value={searchQuery}
                    onChangeText={setSearchQuery}
                    clearButtonMode="while-editing"
                />
                {hasActiveFilters && (
                    <TouchableOpacity onPress={clearFilters} style={styles.clearButton}>
                        <Text style={styles.clearButtonText}>Clear</Text>
                    </TouchableOpacity>
                )}
            </View>

            {/* Tag Filter Chips */}
            {allTags.length > 0 && (
                <View style={styles.tagFilterSection}>
                    <Text style={styles.tagFilterLabel}>Filter by tags:</Text>
                    <ScrollView 
                        horizontal 
                        showsHorizontalScrollIndicator={false}
                        contentContainerStyle={styles.tagFilterContainer}
                    >
                        {allTags.map((tag: string) => (
                            <TouchableOpacity
                                key={tag}
                                style={[
                                    styles.filterTagChip,
                                    selectedTags.includes(tag) && styles.filterTagChipActive
                                ]}
                                onPress={() => toggleTag(tag)}
                            >
                                <Text style={[
                                    styles.filterTagText,
                                    selectedTags.includes(tag) && styles.filterTagTextActive
                                ]}>
                                    {tag}
                                </Text>
                            </TouchableOpacity>
                        ))}
                    </ScrollView>
                </View>
            )}

            {/* Results Count */}
            {hasActiveFilters && (
                <Text style={styles.resultsCount}>
                    {filteredMaterials.length} of {data?.length || 0} materials
                </Text>
            )}

            {error ? (
                <Text style={styles.error}>Failed to load materials</Text>
            ) : (
                <FlatList
                    data={filteredMaterials}
                    renderItem={renderItem}
                    keyExtractor={(item) => item.id}
                    contentContainerStyle={styles.list}
                    refreshControl={
                        <RefreshControl refreshing={isLoading} onRefresh={refetch} />
                    }
                    ListEmptyComponent={
                        <Text style={styles.empty}>
                            {hasActiveFilters 
                                ? 'No materials match your filters'
                                : 'No materials due! Good job.'}
                        </Text>
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
    mainTitle: {
        fontSize: 24,
        fontWeight: 'bold',
        marginBottom: 15,
        color: '#333',
    },
    searchContainer: {
        flexDirection: 'row',
        alignItems: 'center',
        marginBottom: 15,
        gap: 10,
    },
    searchInput: {
        flex: 1,
        backgroundColor: '#fff',
        borderRadius: 8,
        paddingHorizontal: 15,
        paddingVertical: 12,
        fontSize: 16,
        borderWidth: 1,
        borderColor: '#ddd',
    },
    clearButton: {
        backgroundColor: '#d9534f',
        paddingHorizontal: 15,
        paddingVertical: 12,
        borderRadius: 8,
    },
    clearButtonText: {
        color: '#fff',
        fontWeight: '600',
        fontSize: 14,
    },
    tagFilterSection: {
        marginBottom: 15,
    },
    tagFilterLabel: {
        fontSize: 14,
        fontWeight: '600',
        color: '#666',
        marginBottom: 8,
    },
    tagFilterContainer: {
        flexDirection: 'row',
        gap: 8,
        paddingBottom: 5,
    },
    filterTagChip: {
        backgroundColor: '#fff',
        paddingHorizontal: 12,
        paddingVertical: 8,
        borderRadius: 16,
        borderWidth: 1.5,
        borderColor: '#ddd',
    },
    filterTagChipActive: {
        backgroundColor: '#4285F4',
        borderColor: '#4285F4',
    },
    filterTagText: {
        fontSize: 13,
        fontWeight: '600',
        color: '#666',
    },
    filterTagTextActive: {
        color: '#fff',
    },
    resultsCount: {
        fontSize: 13,
        color: '#666',
        marginBottom: 10,
        fontStyle: 'italic',
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
    cardHeader: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 10,
    },
    materialTitle: {
        fontSize: 18,
        fontWeight: '600',
        color: '#333',
        flex: 1,
    },
    badge: {
        backgroundColor: '#4285F4',
        borderRadius: 12,
        paddingHorizontal: 8,
        paddingVertical: 2,
        marginLeft: 10,
    },
    badgeText: {
        color: '#fff',
        fontWeight: 'bold',
        fontSize: 12,
    },
    tagsContainer: {
        flexDirection: 'row',
        flexWrap: 'wrap',
        gap: 8,
    },
    tagBadge: {
        backgroundColor: '#e0e0e0',
        paddingHorizontal: 8,
        paddingVertical: 4,
        borderRadius: 12,
    },
    tagText: {
        fontSize: 12,
        color: '#555',
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
