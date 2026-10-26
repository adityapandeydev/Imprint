# Imprint interface

## Direction
Operate mode: a reader's annotated collection. Typography and the book-cover shelf provide identity; navigation and form controls follow standard web conventions. Taste's editorial-greeting composition informs the centered content grid, adapted to a functional title rather than a greeting. The signature is the cover-first editorial shelf, with a short reading cue beneath each selection.

## Palette
Light: canvas #e5e4e2, surface #e6e0f0, with deep aubergine text and actions. Dark: charcoal-plum canvas #16131b and elevated surfaces #231e2b, stone text #e5e4e2, and pale lavender actions #d3c3e8. The user revised the earlier pale-color reversal to request a genuinely dark counterpart while preserving light mode. Inputs use dark native controls; sage status text and warm gold ratings remain readable. Status hues communicate meaning only.

## Layout and type
Centered 1240px content, 24px grid gaps, 20px mobile gutters. Responsive header uses a separate navigation row on small screens. Newsreader for editorial titles, DM Sans for controls and running text. Cover artwork is contained, never cropped. Compact search results replace the discovery introduction when a query is active.

## Interaction
One save action per discovery book. Collection search, status and shelf filters, and grid/list modes. Secondary card controls expand on demand. Full-page book details and reading statistics. Settings own privacy and data transfer. Dialogs trap and restore focus, close on Escape, and identify their headings. Motion respects the system preference and stays brief.

## Truth and scope
Page changes fade and rise 8px over 220ms with an ease-out curve. Only destination content animates; the shared navigation remains fixed. Search and filter updates do not restart the transition. Reduced-motion preferences disable it.

Editorial selections are static reading starting points, never described as trending or personalized. Statistics use API data. Search can request up to the existing API's 50-result cap; it cannot promise an unlimited catalog. No backend changes in this phase.
