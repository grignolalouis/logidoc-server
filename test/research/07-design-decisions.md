# Design Decisions

## D1: Le JSON est self-contained — pas besoin du PDF après ingestion
- `if_add_node_text = yes` → le texte brut complet est stocké dans chaque node
- Le PDF peut être archivé/supprimé après indexation
- Le JSON est la seule source de vérité pour la retrieval

## D2: Stratégie tables — Parser d'abord, OCR+LLM en fallback
- **Étape 1**: Parser texte classique (pdfcpu ou équivalent Go)
- **Étape 2**: Si une table est mal parsée (détection heuristique ou qualité faible)
  → Convertir la page en image
  → OCR + LLM pour reconstruire la table en texte structuré (markdown table, etc.)
- **Étape 3**: Le texte reconstruit est stocké dans le node JSON comme n'importe quel autre texte
- Pas de focus sur l'optimisation tokens pour l'instant — priorité = robustesse du pipeline

## D3: Priorité = pipeline d'ingestion robuste
- On ne se focus pas sur le coût tokens à ce stade
- L'objectif est un pipeline qui produit un JSON complet et fiable
- Optimisations (modèle moins cher, parallélisation, caching) viendront après
