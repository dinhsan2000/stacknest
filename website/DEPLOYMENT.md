# Deployment Guide - Stacknest Landing Page

Quick guide to deploy the Stacknest landing page to `stacknests.org`.

---

## 🚀 Option 1: GitHub Pages + Custom Domain (Recommended)

Best for: Free hosting, automatic updates on push, CDN included.

### Step 1: Create GitHub Repository

```bash
# Create new repo called "stacknests.org" on GitHub
# Clone it locally
git clone https://github.com/yourusername/stacknests.org
cd stacknests.org
```

### Step 2: Copy Website Files

```bash
# Copy all files from website/ folder
cp -r ../stacknest/website/* ./
git add .
git commit -m "Initial landing page"
git push origin main
```

### Step 3: Enable GitHub Pages

1. Go to **Repository Settings** → **Pages**
2. **Source:** Deploy from branch
3. **Branch:** `main`
4. **Folder:** `/ (root)`
5. **Custom Domain:** `stacknests.org`
6. Click **Save**

### Step 4: Configure DNS

Go to your domain registrar (GoDaddy, Namecheap, etc.):

1. Find **DNS Settings** for `stacknests.org`
2. Add CNAME record:
   - **Type:** CNAME
   - **Name:** `www` (or `@` for root)
   - **Value:** `yourusername.github.io`

3. Add A records for root domain (optional, for apex domain):
   ```
   A record → 185.199.108.153
   A record → 185.199.109.153
   A record → 185.199.110.153
   A record → 185.199.111.153
   ```

GitHub will auto-issue SSL certificate. Wait 5-10 minutes.

✅ Done! Your site is live at `https://stacknests.org`

---

## 🚀 Option 2: Vercel (Even Faster)

Best for: Instant deploys, edge functions support, analytics.

### Step 1: Push to GitHub

```bash
git push origin main
```

### Step 2: Connect Vercel

1. Go to [vercel.com](https://vercel.com)
2. **Import Project** → Select your GitHub repo
3. **Framework:** Other (static site)
4. **Root Directory:** `website`
5. Click **Deploy**

### Step 3: Add Custom Domain

1. In Vercel Dashboard → **Settings** → **Domains**
2. **Add Domain:** `stacknests.org`
3. Follow DNS instructions
4. SSL auto-issued

✅ Done! Deploys automatically on every push.

---

## 🚀 Option 3: Netlify (Drag & Drop)

Best for: Simplicity, form handling, A/B testing.

### Step 1: Prepare Folder

```bash
# Create deployment folder
mkdir stacknests-deploy
cp -r website/* stacknests-deploy/
```

### Step 2: Deploy

1. Go to [netlify.com](https://netlify.com)
2. **Drag & Drop** the `stacknests-deploy` folder
3. Or **Connect Git** and select repo
4. Set **Publish Directory:** `website`

### Step 3: Custom Domain

1. **Site Settings** → **Domain Management**
2. **Add Domain:** `stacknests.org`
3. Update DNS CNAME to Netlify nameservers
4. SSL auto-issued

✅ Done! Your site is live.

---

## 🔄 Continuous Deployment

All three options auto-update when you push to GitHub:

```bash
# Edit the landing page
nano website/index.html

# Commit & push
git add website/
git commit -m "Update landing page content"
git push origin main

# Site updates in 30-60 seconds ✨
```

---

## 📊 Add Analytics

### Google Analytics

Add before `</body>` in `index.html`:

```html
<!-- Google Analytics 4 -->
<script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'G-XXXXXXXXXX');
</script>
```

### Plausible Analytics (Privacy-friendly)

```html
<script defer data-domain="stacknests.org" src="https://plausible.io/js/script.js"></script>
```

### Vercel Analytics (If using Vercel)

Automatic in Vercel dashboard.

---

## 🔒 SSL Certificate

All three options provide **free HTTPS**:
- GitHub Pages: Auto-issued by Let's Encrypt
- Vercel: Auto-issued by DigiCert
- Netlify: Auto-issued by Let's Encrypt

Visit `https://stacknests.org` and check the 🔒 lock icon.

---

## 🎯 Pre-Launch Checklist

- [ ] Download links point to actual releases
- [ ] All screenshots/images are real
- [ ] GitHub repo link is correct
- [ ] Contact/support email is valid
- [ ] FAQ answers are complete
- [ ] No broken links (test with command below)
- [ ] Mobile-responsive (test on phone)
- [ ] Accessibility (test with tab key)
- [ ] Analytics installed and tracking
- [ ] Domain DNS is configured
- [ ] SSL certificate is active

### Test Broken Links

```bash
# Install linkchecker (optional)
pip install linkchecker

# Run check
linkchecker https://stacknests.org
```

---

## 📈 Traffic & Performance

### Check Performance

- **Google PageSpeed Insights:** https://pagespeed.web.dev
- **GTmetrix:** https://gtmetrix.com
- **WebPageTest:** https://www.webpagetest.org

### Monitor Traffic

**GitHub Pages:** Settings → Usage
**Vercel:** Deployment logs & Analytics
**Netlify:** Analytics tab

---

## 🚦 Custom Domain Tips

### Registrar DNS Propagation

After adding CNAME/A records:
- Can take 5 minutes to 48 hours
- Check status: https://www.whatsmydns.net/?q=stacknests.org

### If Domain Doesn't Work

1. Wait 15 minutes
2. Check DNS propagation ☝️
3. Verify CNAME/A records are correct
4. Check platform's DNS instructions

### Common Issues

| Issue | Solution |
|-------|----------|
| **404 Not Found** | Check publish folder is correct |
| **HTTPS fails** | Wait for cert issuance (5-10 min) |
| **DNS not resolving** | Wait for propagation, verify records |
| **Slow load times** | Use CDN (all platforms provide this) |

---

## 🆘 Troubleshooting

### GitHub Pages 404

```bash
# Check file structure
ls -la website/
# Should see: index.html, README.md, .gitignore

# Verify it's committed
git log --oneline | head -5
```

### Vercel Build Fails

1. Check build logs in Vercel dashboard
2. Ensure `website` folder has `index.html` at root
3. No npm dependencies needed (pure HTML/CSS)

### Domain Doesn't Resolve

1. Check DNS with: `nslookup stacknests.org`
2. Verify CNAME points to your hosting
3. Wait for DNS propagation (up to 48 hours)

---

## 🔄 Update Workflow

Typical update process:

```bash
# 1. Make changes locally
nano website/index.html

# 2. Test in browser
# (Open file:// in browser or use Python server)
python -m http.server 8000
# Visit http://localhost:8000/website/

# 3. Commit & push
git add website/
git commit -m "Update: [what changed]"
git push origin main

# 4. Verify deployment (2-3 minutes)
curl https://stacknests.org | grep "Stacknest"
```

---

## 📞 Support

Having issues? Try:

1. Check platform docs (GitHub/Vercel/Netlify)
2. Verify DNS with tools like `whatsmydns.net`
3. Check build/deploy logs in platform dashboard
4. Search for error message on platform forums

---

**Happy deploying! 🚀**
