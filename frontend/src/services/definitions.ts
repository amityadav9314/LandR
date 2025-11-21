import { AuthServiceImplementation, LoginRequest, LoginResponse } from '../../proto/auth';
import { LearningServiceImplementation, AddMaterialRequest, AddMaterialResponse, FlashcardList, CompleteReviewRequest } from '../../proto/learning';
import { Empty } from '../../google/protobuf/empty';

export const AuthServiceDefinition = {
    name: 'AuthService',
    fullName: 'auth.AuthService',
    methods: {
        login: {
            name: 'Login',
            requestType: LoginRequest,
            requestStream: false,
            responseType: LoginResponse,
            responseStream: false,
            options: {},
        },
    },
} as const;

export const LearningServiceDefinition = {
    name: 'LearningService',
    fullName: 'learning.LearningService',
    methods: {
        addMaterial: {
            name: 'AddMaterial',
            requestType: AddMaterialRequest,
            requestStream: false,
            responseType: AddMaterialResponse,
            responseStream: false,
            options: {},
        },
        getDueFlashcards: {
            name: 'GetDueFlashcards',
            requestType: Empty,
            requestStream: false,
            responseType: FlashcardList,
            responseStream: false,
            options: {},
        },
        completeReview: {
            name: 'CompleteReview',
            requestType: CompleteReviewRequest,
            requestStream: false,
            responseType: Empty,
            responseStream: false,
            options: {},
        },
    },
} as const;
