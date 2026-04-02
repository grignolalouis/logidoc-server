# PageIndex Ingestion — Deep Dive (Q&A)

## Q1: Comment sont stockés les documents ?

### L'index (output de l'ingestion)
Stocké en **JSON** dans `./results/{filename}_structure.json`:
```json
{
  "doc_name": "report.pdf",
  "doc_description": "Annual financial report...",
  "structure": [
    {
      "title": "Chapter 1",
      "node_id": "0001",
      "start_index": 1,
      "end_index": 15,
      "summary": "Covers revenue and growth metrics.",
      "text": "Full extracted text of this section...",
      "nodes": [/* children */]
    }
  ]
}
```

### Le document source
Le PDF/Markdown original est conservé tel quel. L'index JSON **pointe** vers les pages
du document source via `start_index`/`end_index`. Pendant la retrieval, le système
retourne aux pages originales pour extraire le texte brut.

### Approches de stockage possibles (pour notre intégration Go):
1. **Filesystem** — JSON files, le plus simple (ce que fait PageIndex Python)
2. **Base de données KV** — Redis, BoltDB pour du lookup rapide par doc_id
3. **SQL** — PostgreSQL/SQLite pour du multi-tenant avec metadata search
4. **Object storage** — S3/MinIO pour les documents sources volumineux

---

## Q2: Le parsing — Comment on extrait le texte ?

### PageIndex utilise 2 parsers:
1. **PyMuPDF (fitz)** — parser principal, extrait texte + images par page
2. **PyPDF2** — fallback

### La fonction clé: `get_page_tokens()`
- Itère page par page sur le PDF
- Extrait le texte brut de chaque page
- Compte les tokens avec `tiktoken.encoding_for_model()`
- Retourne une liste de tuples `(page_text, token_count)`

### Pour notre intégration Go:
- **pdfcpu** — bibliothèque Go native pour PDF
- **unipdf** — plus complet mais commercial
- **Appel externe** — `pdftotext` (poppler), ou un service de parsing

---

## Q3: Les tables — Comment les traiter ?

### PageIndex a 2 approches:

#### Approche 1: Text-based (par défaut)
- PyMuPDF extrait les tables comme du texte brut
- La structure tabulaire est **partiellement perdue** (colonnes deviennent du texte linéaire)
- Le LLM doit reconstituer la structure à partir du contexte
- **Limitation reconnue** — pas de traitement spécifique des tables

#### Approche 2: Vision-based RAG (cookbook séparé)
- **Pas d'OCR du tout** — les pages sont converties en images (PyMuPDF à 2x zoom)
- Les images de pages sont envoyées directement à un **VLM** (GPT-4.1, etc.)
- Le VLM interprète visuellement tables, charts, figures, formules
- Pipeline: PDF → page images → tree navigation → VLM lit l'image directement

#### Pour notre intégration:
- **Option 1**: Parser texte classique + laisser le LLM comprendre
- **Option 2**: Vision mode — convertir pages en images, envoyer au VLM
- **Option 3**: Parser spécialisé tables (camelot, tabula) en pré-processing
- **Option 4**: Utiliser un OCR avancé (Mistral OCR, Gemini) comme étape de parsing ///////

---

## Q4: L'ingestion en détail — Les 3 modes

Le cœur de l'ingestion est `meta_processor()` qui dispatche vers 3 modes:

### Mode 1: TOC avec numéros de page
```
PDF → scan 20 premières pages → TOC trouvée avec "Chapter 1 ......... p.5"
  → LLM extrait la structure hiérarchique
  → LLM vérifie que les titres apparaissent bien aux pages indiquées
  → Si accuracy < 60% → fallback vers Mode 2 ou 3
  → Sinon → tree construit avec page ranges précis
```

### Mode 2: TOC sans numéros de page
```
PDF → TOC trouvée mais sans numéros ("Chapter 1: Introduction")
  → LLM extrait la liste de sections
  → Pour chaque section, LLM cherche la page de début dans le document
  → Vérification + correction
  → Tree construit
```

### Mode 3: Pas de TOC (`process_no_toc`)
```
PDF → Aucune TOC détectée
  → Le texte complet est envoyé au LLM en morceaux
  → generate_toc_init(): premier chunk → structure initiale (1, 1.1, 1.2...)
  → generate_toc_continue(): chunks suivants → continuation de la structure
  → Les morceaux sont envoyés séquentiellement au LLM
  → Tree construit à partir de la structure générée
```

### Post-processing (commun aux 3 modes):
1. `post_processing()` — calcule les `end_index` de chaque node
2. `list_to_tree()` — convertit la liste plate en arbre hiérarchique
3. `write_node_id()` — assigne des IDs zero-padded ("0001", "0002"...)
4. `process_large_node_recursively()` — subdivise les gros nodes
5. Génération de summaries (optionnel) — LLM résume chaque node
6. Génération de description du document (optionnel)

---

## Q5: Réutiliser un document déjà indexé

### Dans PageIndex Python:
- L'index est sauvé en `./results/{name}_structure.json`
- Pour réutiliser: **charger le JSON**, c'est tout
- Pas de mécanisme "catalogue" intégré — c'est juste des fichiers JSON
- L'API cloud (PageIndex SaaS) gère un catalogue avec folders/workspaces

### Ce qu'il faut construire pour notre intégration:
```
DocumentStore (interface)
  ├── RegisterDocument(doc) → docID
  ├── GetIndex(docID) → tree JSON
  ├── ListDocuments() → []DocumentMeta
  ├── DeleteDocument(docID)
  └── GetDocumentPages(docID, startPage, endPage) → text
```

L'agent doit pouvoir:
1. Lister les documents disponibles (déjà indexés)
2. Sélectionner un document par nom/ID/metadata
3. Charger son index sans re-indexer
4. Indexer un nouveau document si nécessaire

---

## Q6: Document de 400 pages — Comment ça marche ?

### Étape 1: Parsing
- 400 pages × ~500 tokens/page = ~200,000 tokens de texte brut
- Chaque page est parsée individuellement → liste de 400 tuples (text, tokens)

### Étape 2: Détection TOC
- Scan des 20 premières pages (configurable via `--toc-check-pages`)
- LLM classifie chaque page: "est-ce une page de TOC ?"
- Coût: ~20 appels LLM légers (quelques centaines de tokens chacun)

### Étape 3a: SI une TOC existe
- Le LLM traite UNIQUEMENT les pages de TOC (quelques pages)
- Extrait la structure hiérarchique → peut-être 30-50 sections
- Vérifie par sampling que les titres correspondent aux bonnes pages
- **Très efficace** — seules les pages TOC sont envoyées au LLM

### Étape 3b: SI pas de TOC (le cas coûteux)
- Le texte est découpé en groupes de pages
- `generate_toc_init()` reçoit le premier groupe
- `generate_toc_continue()` reçoit les groupes suivants
- **Le document entier passe par le LLM**, mais en morceaux séquentiels
- Pour 400 pages: potentiellement 40+ appels LLM (groupes de ~10 pages)
- Chaque appel: ~5,000-10,000 tokens input

### Étape 4: Subdivision des gros nodes
- Un node "Chapter 3" couvrant 80 pages dépasse `max_page_num_each_node` (10)
- `process_large_node_recursively()` extrait ces 80 pages
- Les envoie au LLM en mode `process_no_toc` pour créer des sous-sections
- Récursif jusqu'à ce que chaque node < 10 pages ET < 20,000 tokens

### Estimation tokens pour 400 pages sans TOC:
- Input: ~200,000 tokens (tout le document, en morceaux)
- Output: ~10,000-20,000 tokens (structure JSON générée)
- + Subdivision: ~50,000-100,000 tokens additionnels pour les gros nodes
- + Summaries: ~50,000 tokens si activés
- **Total estimé: 300,000 - 400,000 tokens pour l'indexation**
- **Coût avec GPT-4o: ~$1-3 par document de 400 pages**

---

## Q7: Token consumption — Est-ce normal ?

### Oui, c'est le trade-off fondamental de PageIndex:

**Investissement upfront (indexation):**
- 300K-400K tokens pour un doc de 400 pages SANS TOC
- Beaucoup moins (~20K-50K) si le doc A une TOC
- C'est un coût **one-time** par document

**Économie à la retrieval (chaque query):**
- L'index complet (titres + summaries) tient en ~5,000-10,000 tokens
- Navigation: ~2,000 tokens (query + TOC compacte)
- Extraction: 0 tokens (pure code operation)
- Réponse: ~2,000-5,000 tokens (petit extrait + query)
- **Total par query: ~5,000-15,000 tokens** au lieu de 200,000

### Comparaison:
| Approche           | Indexation          | Par query           |
|--------------------|---------------------|---------------------|
| Pas d'index        | 0 tokens            | 200,000 tokens      |
| Vector RAG         | ~200,000 (embeddings) | ~5,000-10,000     |
| PageIndex avec TOC | ~20,000-50,000      | ~5,000-15,000       |
| PageIndex sans TOC | ~300,000-400,000    | ~5,000-15,000       |

**Conclusion**: Le coût d'indexation est amorti dès la 2ème-3ème query.
Pour un doc de 400 pages consulté 10 fois, PageIndex est nettement plus
économique que d'envoyer le doc complet à chaque fois.

### Optimisations possibles:
1. **Utiliser un modèle moins cher** pour l'indexation (GPT-4o-mini, Haiku)
2. **Prompt caching** — si le même document est re-indexé
3. **Détection de structure heuristique** avant le LLM (regex sur headers)
4. **Map-reduce parallèle** — traiter les chunks en parallèle, pas séquentiellement
5. **Indexation incrémentale** — ne re-indexer que les sections modifiées
