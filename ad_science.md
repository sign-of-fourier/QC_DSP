# Ad Science
*Traditional ad set optimization assumes independence (adding ads always helps) or uses naive A/B testing. We use a Bayesian framework with embedding-aware diversity that automatically balances portfolio construction—maximizing total lift while preventing redundant creative cannibalization.*

## The Core Problem

**If conversions were truly additive**, the objective would be:

$$
\text{maximize } \mathbb{E}[\text{conversions}_{ad1}] + \mathbb{E}[\text{conversions}_{ad2}] + \ldots + \mathbb{E}[\text{conversions}_{adn}]
$$

In this case, you'd independently select the ads with highest expected conversions. PI (probability of improvement) would be **inappropriate** because you're not looking for "the one best ad"—you want the sum.

**But cannibalization implies substitution effects**, meaning:

- User would have clicked ad1 OR ad2, not both
- The conversion outcome is non-additive
- The true model is something like: $\mathbb{E}[\max(\text{conv}_1, \text{conv}_2, \ldots)]$ or $\mathbb{E}[\text{any ad converts}]$


## Why Embedding Orthogonality Matters

Your intuition about pushing ads to be orthogonal in embedding space is trying to **approximate additivity by minimizing overlap**. If ads are sufficiently different (orthogonal embeddings), they appeal to different user subpopulations, reducing substitution effects.

This is actually a **hybrid objective**:

$$
\text{maximize } \sum_i \mathbb{E}[\text{conv}_i] - \lambda \cdot \text{Overlap}(ad_i, ad_j)
$$

Where the overlap term captures cannibalization. The embedding distance in GPR space is acting as a proxy for this overlap term.

## What's Actually Happening in Practice

The confusion arises because **cannibalization isn't binary**—it exists on a spectrum:

1. **Pure substitution** (100% cannibalization): User clicks exactly one ad, doesn't matter which. Objective: $\mathbb{P}(\text{any conversion})$. Use batch PI or exceedance probability.
2. **Pure independence** (0% cannibalization): Each ad reaches distinct users. Objective: $\sum \mathbb{E}[\text{conv}_i]$. Use expected improvement sum.
3. **Partial cannibalization** (reality): Some overlap, some independence. Objective: Something in between that depends on embedding distance.

## The Mathematical Framework You Need

For partial cannibalization, you need a **copula-based model** or an **interaction model** in your GPR:

**Option 1: Model the joint distribution**
Instead of independent GPs for each ad, model:

$$
f(\text{ad}_1, \text{ad}_2, \ldots, \text{ad}_n \mid \text{context}) \sim \mathcal{GP}(\mu, K)
$$

Where the kernel $K$ includes **cross-ad covariance terms** that capture cannibalization. When embeddings are similar, the covariance is high (high cannibalization). When orthogonal, covariance → 0 (independence).

**Option 2: Submodular optimization**
Model conversions as a **submodular set function**, where adding similar ads has diminishing marginal returns. The acquisition function becomes:[^1]

$$
\text{maximize } f(S) \text{ where } f(S \cup \{ad\}) - f(S) \text{ decreases as } |S| \text{ grows}
$$

This naturally captures cannibalization—the first ad gives full lift, similar ads add less.

## Practical Decision for Your Platform

**Short-term pragmatic approach:**
Use batch PI with embedding diversity as you're doing, but acknowledge you're implicitly assuming **moderate-to-high substitution**. This is defensible for most ad sets where users see multiple ads but only click one.

**Medium-term enhancement:**
Add an **empirical calibration layer**:

- Measure actual additivity vs. substitution for different embedding distances from historical data
- Build a correction factor: $\text{expected total conversions} = \sum_i \mathbb{E}[\text{conv}_i] \cdot (1 - c_{ij})$ where $c_{ij}$ is empirical cannibalization between ads $i$ and $j$
- Use this to weight the acquisition function

**Long-term research direction:**
Implement multivariate GPR with learned cross-ad kernels that capture cannibalization structure directly. This is theoretically cleaner but much harder to implement and requires more data.

<div align="center">⁂</div>

[^1]: http://proceedings.mlr.press/v32/gopalan14.pdf

[^2]: https://pixis.ai/blog/8-ways-to-avoid-ad-cannibalization-in-paid-media-campaigns/

[^3]: https://www.reddit.com/r/PPC/comments/7yo9cf/quantifying_cannibalization_of_organic_traffic/

[^4]: https://www.topsort.com/post/how-to-solve-the-problem-of-ad-cannibalization-in-retail-media

[^5]: https://www.moburst.com/blog/what-is-cannibalization-in-aso/

[^6]: https://searchengineland.com/prevent-ppc-cannibalizing-seo-efforts-451920

[^7]: https://www.practicalecommerce.com/understanding-google-ads-new-conversion-action-sets

[^8]: http://math.iit.edu/~mdixon7/multivariate-gaussian-process_DC.pdf

[^9]: https://www.branch.io/resources/blog/apple-search-ads-strategies-for-managing-cannibalization-and-competition/

[^10]: https://www.reddit.com/r/FacebookAds/comments/xk0p6c/do_ad_sets_within_same_campaign_affect_the/

[^11]: https://arxiv.org/html/2212.01048v2

[^12]: https://en.wikipedia.org/wiki/Multi-armed_bandit

[^13]: https://deeprlcourse.github.io/recitations/week6/

[^14]: https://arxiv.org/html/2505.13355v2

[^15]: https://www.economics.uci.edu/files/docs/micro/s11/Scott.pdf

[^16]: https://naomi.princeton.edu/wp-content/uploads/sites/744/2021/03/Allerton2013ol.pdf

[^17]: http://papers.neurips.cc/paper/7179-action-centered-contextual-bandits.pdf

[^18]: https://www.sciencedirect.com/science/article/abs/pii/S0377221724007203

[^19]: https://www.countbayesie.com/blog/2020/9/26/learn-thompson-sampling-by-building-an-ad-auction

[^20]: http://proceedings.mlr.press/v39/chou14.pdf
