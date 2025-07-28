# Secretly Web Dashboard

A modern, responsive web interface for the Secretly secret management system.

## Features

- 🔐 Secure authentication and session management
- 📱 Responsive design for desktop and mobile
- 🌍 Internationalization support (EN, RU, ES, FR, DE)
- ♿ Accessibility compliant (WCAG 2.1 AA)
- 🎨 Modern UI with Tailwind CSS
- ⚡ Fast development with Vite
- 🧪 Comprehensive testing with Vitest and Playwright
- 📊 State management with Zustand and React Query

## Tech Stack

- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **UI Components**: Headless UI
- **State Management**: Zustand + React Query
- **Forms**: React Hook Form + Zod validation
- **Routing**: React Router v6
- **Testing**: Vitest + React Testing Library + Playwright
- **Internationalization**: React i18next

## Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── ui/             # Basic UI components (Button, Input, etc.)
│   ├── forms/          # Form components
│   ├── layout/         # Layout components (Header, Sidebar, etc.)
│   └── features/       # Feature-specific components
├── pages/              # Page components
│   ├── auth/           # Authentication pages
│   ├── dashboard/      # Dashboard pages
│   ├── secrets/        # Secret management pages
│   ├── sharing/        # Sharing management pages
│   ├── profile/        # User profile pages
│   └── admin/          # Administrative pages
├── hooks/              # Custom React hooks
├── services/           # API client and services
├── store/              # State management (Zustand stores)
├── types/              # TypeScript type definitions
├── utils/              # Utility functions
├── i18n/               # Internationalization setup
└── test/               # Test utilities and setup
```

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Installation

1. Install dependencies:
   ```bash
   npm install
   ```

2. Copy environment configuration:
   ```bash
   cp .env.example .env.development
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

The application will be available at `http://localhost:3000`.

## Available Scripts

### Development
- `npm run dev` - Start development server
- `npm run build:dev` - Build for development
- `npm run preview` - Preview production build locally

### Production
- `npm run build` - Build for production
- `npm run build:prod` - Build for production (explicit)
- `npm run build:analyze` - Build and analyze bundle size

### Code Quality
- `npm run lint` - Run ESLint
- `npm run lint:fix` - Fix ESLint issues
- `npm run format` - Format code with Prettier
- `npm run format:check` - Check code formatting
- `npm run type-check` - Run TypeScript type checking

### Testing
- `npm run test` - Run unit tests in watch mode
- `npm run test:run` - Run unit tests once
- `npm run test:coverage` - Run tests with coverage
- `npm run test:ui` - Run tests with UI
- `npm run e2e` - Run end-to-end tests
- `npm run e2e:ui` - Run E2E tests with UI
- `npm run e2e:debug` - Debug E2E tests

### Utilities
- `npm run clean` - Clean build artifacts
- `npm run prepare` - Run pre-commit checks

## Environment Variables

Create a `.env.development` file based on `.env.example`:

```env
# API Configuration
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT=30000

# Application Configuration
VITE_APP_NAME=Secretly
VITE_APP_VERSION=1.0.0
VITE_ENVIRONMENT=development

# Feature Flags
VITE_ENABLE_DEBUG=true
VITE_ENABLE_DEVTOOLS=true

# Security
VITE_SESSION_TIMEOUT=3600000
VITE_CLIPBOARD_CLEAR_TIMEOUT=30000

# UI Configuration
VITE_DEFAULT_LANGUAGE=en
VITE_DEFAULT_THEME=system
VITE_ITEMS_PER_PAGE=20
```

## API Integration

The web dashboard integrates with the existing Secretly HTTP REST API:

- **Base URL**: Configurable via `VITE_API_BASE_URL`
- **Authentication**: JWT tokens with HTTP-only cookies
- **CORS**: Configured for cross-origin requests
- **Error Handling**: Centralized error handling with user-friendly messages

## Internationalization

The application supports multiple languages:

- English (en) - Default
- Russian (ru)
- Spanish (es)
- French (fr)
- German (de)

Language files are located in `src/i18n/locales/` and integrate with the existing backend translation system.

## Accessibility

The application follows WCAG 2.1 AA guidelines:

- Semantic HTML structure
- ARIA labels and descriptions
- Keyboard navigation support
- Screen reader compatibility
- High contrast mode support
- Focus management

## Security Features

- HTTPS enforcement
- CSRF protection
- XSS prevention through input sanitization
- Secure session management
- Automatic clipboard clearing
- Content Security Policy (CSP)

## Performance Optimizations

- Code splitting and lazy loading
- Bundle optimization with manual chunks
- Image optimization and lazy loading
- Service worker for caching (future enhancement)
- React Query for intelligent data caching

## Testing Strategy

### Unit Tests
- Component testing with React Testing Library
- Hook testing with custom utilities
- Utility function testing
- State management testing

### Integration Tests
- API integration testing
- Form submission workflows
- Authentication flows
- Navigation testing

### End-to-End Tests
- Complete user workflows
- Cross-browser testing
- Responsive design testing
- Accessibility testing

## Contributing

1. Follow the established project structure
2. Write tests for new features
3. Ensure accessibility compliance
4. Follow TypeScript strict mode
5. Use semantic commit messages
6. Run pre-commit checks before submitting

## Deployment

The web dashboard is designed to be served as static assets from the existing Go HTTP server:

1. Build the application: `npm run build:prod`
2. Copy `dist/` contents to the Go server's static assets directory
3. Configure the Go server to serve static files
4. Set up proper HTTP headers and security policies

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## License

This project is part of the Secretly secret management system.