# Stacknest Landing Page

Official landing page for Stacknest - Local Development Environment Manager.

**Live at:** https://stacknests.org

---

## 📁 Project Structure

```
website/
├── index.html          # Main landing page (HTML + Tailwind CSS)
├── README.md           # This file
└── .gitignore          # Git ignore rules
```

---

## 🎨 Design System

- **Pattern:** App Store Style Landing
- **Style:** Vibrant & Block-based (Bold, energetic, modern)
- **Colors:**
  - Primary: `#3B82F6` (Blue)
  - Secondary: `#1E293B` (Slate)
  - Background: `#0F172A` (Dark Slate)
  - Text: `#F1F5F9` (Off-white)

- **Typography:**
  - Display: Space Grotesk (700 weight)
  - Body: DM Sans (400, 500, 700 weights)
  - From Google Fonts (free)

- **Tech Stack:**
  - HTML5 (semantic)
  - Tailwind CSS v3 (via CDN)
  - Vanilla JavaScript (no dependencies)

---

## ⚡ Features

✅ **SEO-Optimized**
- Semantic HTML5 structure
- Meta tags for social sharing
- Proper heading hierarchy
- Mobile-first responsive design

✅ **Accessibility**
- WCAG AA compliant contrast ratios
- Focus states for keyboard navigation
- Alt text for all images
- `prefers-reduced-motion` support

✅ **Performance**
- Single HTML file (no build step)
- Tailwind CSS via CDN
- Optimized images
- No JavaScript dependencies

✅ **Responsive**
- Mobile: 375px
- Tablet: 768px
- Desktop: 1024px, 1440px
- Dark theme optimized

---

## 🚀 Deployment

### Option 1: GitHub Pages (Recommended)

1. Create a `stacknests.github.io` repository
2. Push this `website/` folder to the repo
3. GitHub will automatically serve it at `https://stacknests.github.io`
4. Point `stacknests.org` domain to GitHub Pages via DNS CNAME

**GitHub Pages Setup:**
```bash
# In repo settings → Pages
# Source: Deploy from branch
# Branch: main
# Folder: /root
```

### Option 2: Vercel (Free, fast)

1. Connect GitHub repo to Vercel
2. Set root directory: `website`
3. Deploy (automatic on push)
4. Add custom domain `stacknests.org`

### Option 3: Netlify (Free, easy)

1. Drag & drop the `website` folder
2. Or connect GitHub repo
3. Set publish directory: `website`
4. Add custom domain

---

## 📝 Editing the Page

### Change Text Content

Edit HTML directly in `index.html`. No build step needed.

Common sections to update:
- `<h1>` in Hero section
- Feature cards (change text/icons)
- Download buttons (add actual links)
- FAQ items
- Footer links

### Change Colors

Edit the `:root` CSS variables at top of `<style>` tag:

```css
:root {
    --primary: #3B82F6;        /* Change this */
    --primary-dark: #2563EB;
    --secondary: #1E293B;
    --bg: #0F172A;
    --text: #F1F5F9;
}
```

### Change Fonts

Edit the Tailwind config in `<script>` tag:

```javascript
fontFamily: {
    sans: ['DM Sans', 'sans-serif'],    /* Body font */
    display: ['Space Grotesk', 'sans-serif'],  /* Headings */
}
```

### Add Images

Replace placeholder divs with actual `<img>` tags:

```html
<img 
    src="https://images.unsplash.com/photo-..."
    alt="Descriptive alt text for SEO"
    class="rounded-lg"
>
```

---

## ✅ Pre-Launch Checklist

- [ ] Update all download links to actual binary URLs
- [ ] Add real screenshots of Stacknest app
- [ ] Update GitHub repo link in footer
- [ ] Change all `#` href links to real URLs
- [ ] Test responsive design on mobile
- [ ] Test accessibility with keyboard navigation
- [ ] Test dark/light mode (if using system preference)
- [ ] Update Open Graph meta tags for social sharing
- [ ] Set up analytics (Google Analytics, Plausible, etc.)
- [ ] Test form submissions (if adding contact form)

---

## 🔍 SEO Tips

1. **Keywords:** Developer tools, local environment, Apache, Nginx, MySQL, PHP
2. **Meta Description:** Already set in `<meta name="description">`
3. **Open Graph:** Add social sharing tags:
   ```html
   <meta property="og:title" content="Stacknest">
   <meta property="og:description" content="...">
   <meta property="og:image" content="...">
   ```
4. **Structured Data:** Add JSON-LD schema for rich snippets
5. **Mobile:** Responsive + mobile-first indexing ready

---

## 📊 Analytics

Add analytics script before `</body>`:

```html
<!-- Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_MEASUREMENT_ID');
</script>
```

---

## 🛠️ Customization Examples

### Add Newsletter Signup

```html
<form method="POST" action="https://your-service.com/subscribe">
    <input type="email" name="email" placeholder="your@email.com" required>
    <button type="submit" class="btn-primary">Subscribe</button>
</form>
```

### Add Testimonials

```html
<div class="feature-card">
    <p class="italic text-slate-300">"Stacknest changed my dev workflow completely!"</p>
    <p class="font-bold mt-4">— John Developer</p>
</div>
```

### Add Video Demo

```html
<iframe 
    width="100%" 
    height="400" 
    src="https://www.youtube.com/embed/VIDEO_ID"
    title="Stacknest Demo"
    frameborder="0" 
    allowfullscreen>
</iframe>
```

---

## 📄 License

This landing page is part of Stacknest, licensed under MIT.

---

## 🤝 Contributing

To contribute improvements to this landing page:

1. Clone the repo
2. Edit `website/index.html`
3. Test locally in browser
4. Submit PR with description

---

## 📞 Support

- **Issues:** GitHub Issues
- **Discussions:** GitHub Discussions
- **Email:** dinhsan200@gmail.com

---

**Built with ❤️ using HTML5, Tailwind CSS, and vanilla JavaScript.**
