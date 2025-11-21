import { createChannel, createClientFactory, ClientError, Status, Metadata } from 'nice-grpc-web';
import { Platform } from 'react-native';
import { AuthServiceClient } from '../../proto/auth';
import { LearningServiceClient } from '../../proto/learning';
import { AuthServiceDefinition, LearningServiceDefinition } from './definitions';
import { API_URL } from '../utils/config';

// Platform-aware storage helper
const getToken = async (): Promise<string | null> => {
    if (Platform.OS === 'web') {
        return localStorage.getItem('auth_token');
    } else {
        const SecureStore = await import('expo-secure-store');
        return await SecureStore.getItemAsync('auth_token');
    }
};

// Create the gRPC-Web channel
const channel = createChannel(API_URL);

// Create a client factory
const clientFactory = createClientFactory().use(async function* (call, options) {
    // Auth Interceptor
    const token = await getToken();

    const metadata = new Metadata(options.metadata);
    if (token) {
        metadata.set('authorization', `Bearer ${token}`);
    }

    try {
        const generator = call.next(call.request, { ...options, metadata });
        for await (const response of generator) {
            yield response;
        }
    } catch (error) {
        if (error instanceof ClientError && error.code === Status.UNAUTHENTICATED) {
            // Handle unauthenticated error (e.g., logout user)
            console.log("User unauthenticated");
        }
        throw error;
    }
});

// Export initialized clients
export const authClient: AuthServiceClient = clientFactory.create(
    AuthServiceDefinition,
    channel
);

export const learningClient: LearningServiceClient = clientFactory.create(
    LearningServiceDefinition,
    channel
);
