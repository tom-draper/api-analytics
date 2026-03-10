<?php

declare(strict_types=1);

namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\Routing\Attribute\Route;

/**
 * Example controller for Symfony API Analytics.
 *
 * Setup:
 * 1. Add API_ANALYTICS_KEY=your-api-key-here to your .env file
 * 2. Register the event subscriber in config/services.yaml (see below)
 * 3. Define your routes as normal - all requests are logged automatically
 *
 * config/services.yaml:
 *
 *     ApiAnalytics\Symfony\AnalyticsEventSubscriber:
 *         arguments:
 *             $apiKey: '%env(API_ANALYTICS_KEY)%'
 *         tags:
 *             - { name: kernel.event_subscriber }
 */
class ApiController extends AbstractController
{
    #[Route('/', methods: ['GET'])]
    public function index(): JsonResponse
    {
        return $this->json(['message' => 'Hello, World!']);
    }

    #[Route('/health', methods: ['GET'])]
    public function health(): JsonResponse
    {
        return $this->json(['status' => 'healthy']);
    }

    #[Route('/users', methods: ['GET'])]
    public function users(): JsonResponse
    {
        return $this->json([
            'users' => [
                ['id' => 1, 'name' => 'Alice'],
                ['id' => 2, 'name' => 'Bob'],
            ],
        ]);
    }

    #[Route('/users/{id}', methods: ['GET'])]
    public function user(int $id): JsonResponse
    {
        return $this->json(['id' => $id, 'name' => 'Alice']);
    }

    #[Route('/users', methods: ['POST'])]
    public function createUser(Request $request): JsonResponse
    {
        $data = json_decode($request->getContent(), true);

        return $this->json(['id' => 3, 'name' => $data['name'] ?? ''], 201);
    }
}
