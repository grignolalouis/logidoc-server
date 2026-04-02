# Guide d'Architecture Hexagonale

## Introduction

L'architecture hexagonale, aussi appelée "Ports & Adapters", est un pattern architectural qui isole la logique métier des détails techniques comme les bases de données, les APIs et les interfaces utilisateur.

Le principe fondamental est que le code métier ne dépend de rien d'externe. Les dépendances pointent toujours vers l'intérieur.

## Les trois zones

### Le Core (domaine)

Le core contient les entités métier, les value objects et les règles de gestion. Il ne connaît ni HTTP, ni base de données, ni framework. C'est le coeur de l'application.

Exemple: un `Document` avec un statut qui passe de `pending` à `ready` après indexation.

### Les Ports (interfaces)

Les ports sont des interfaces Go qui définissent ce que le core attend ou expose:
- **Ports primaires**: ce que l'extérieur peut demander (DocumentService, RetrievalService)
- **Ports secondaires**: ce dont le core a besoin (DocumentRepository, LLMProvider)

### Les Adapters (implémentations)

Les adapters implémentent les ports:
- **Primary adapters**: HTTP handlers (Fiber), MCP server (trpc-mcp-go)
- **Secondary adapters**: MongoDB repos, LLM adapter (OpenAI/Anthropic)

## Avantages

1. **Testabilité**: on mock les ports pour tester le core sans infrastructure
2. **Flexibilité**: changer de base de données = changer un adapter, pas le core
3. **Clarté**: la séparation des responsabilités est explicite dans la structure

## Quand l'utiliser

L'architecture hexagonale est adaptée pour:
- Les projets avec une logique métier complexe
- Les systèmes qui doivent supporter plusieurs interfaces (API REST, MCP, CLI)
- Les applications qui évoluent fréquemment côté infrastructure

Elle est excessive pour un simple CRUD sans logique métier.
