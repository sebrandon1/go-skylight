# Rewards

## List rewards and star balances

```bash
skylight reward list
skylight reward points
```

## Create a reward

```bash
skylight reward create --title "Movie night pick" --points 50 --emoji-icon 🎬
```

## Redeem a reward for a family member

```bash
# Get reward ID from `reward list`, category ID from `category list`
skylight reward redeem --reward-id REWARD_ID --category-id CATEGORY_ID
```

## Undo a redemption

```bash
skylight reward unredeem --reward-id REWARD_ID --category-id CATEGORY_ID
```

## Update a reward

```bash
skylight reward update --reward-id REWARD_ID --title "Pizza pick" --points 30
```

## Delete a reward

```bash
skylight reward delete --reward-id REWARD_ID
```
