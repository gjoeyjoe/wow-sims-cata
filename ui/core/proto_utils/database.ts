import {
	EquipmentSpec,
	GemColor,
	ItemSlot,
	ItemSpec,
} from '../proto/common.js';
import {
	IconData,
	UIDatabase,
	UIEnchant as Enchant,
	UIGem as Gem,
	UIItem as Item,
} from '../proto/ui.js';

import {
	getEligibleEnchantSlots,
	getEligibleItemSlots,
} from './utils.js';
import { gemEligibleForSocket, gemMatchesSocket } from './gems.js';
import { EquippedItem } from './equipped_item.js';
import { Gear } from './gear.js';

const dbUrlJson = '/wotlk/assets/database/db.json';
const dbUrlBin = '/wotlk/assets/database/db.bin';
const READ_JSON = false;

export class Database {
	private static loadPromise: Promise<Database>|null = null;
	static get(): Promise<Database> {
		if (Database.loadPromise == null) {
			if (READ_JSON) {
				Database.loadPromise = fetch(dbUrlJson)
					.then(response => response.json())
					.then(json => new Database(UIDatabase.fromJson(json)));
			} else {
				Database.loadPromise = fetch(dbUrlBin)
					.then(response => response.arrayBuffer())
					.then(buffer => new Database(UIDatabase.fromBinary(new Uint8Array(buffer))));
			}
		}
		return Database.loadPromise;
	}

	private readonly items: Record<number, Item> = {};
	private readonly enchantsBySlot: Partial<Record<ItemSlot, Enchant[]>> = {};
	private readonly gems: Record<number, Gem> = {};
	private readonly itemIcons: Record<number, IconData>;
	private readonly spellIcons: Record<number, IconData>;

	private constructor(db: UIDatabase) {
		db.items.forEach(item => this.items[item.id] = item);
		db.enchants.forEach(enchant => {
			const slots = getEligibleEnchantSlots(enchant);
			slots.forEach(slot => {
				if (!this.enchantsBySlot[slot]) {
					this.enchantsBySlot[slot] = [];
				}
				this.enchantsBySlot[slot]!.push(enchant);
			});
		});
		db.gems.forEach(gem => this.gems[gem.id] = gem);

		this.itemIcons = {};
		this.spellIcons = {};
		db.itemIcons.forEach(data => this.itemIcons[data.id] = data);
		db.spellIcons.forEach(data => this.spellIcons[data.id] = data);
	}

	getItems(slot: ItemSlot): Array<Item> {
		let items = Object.values(this.items);
		items = items.filter(item => getEligibleItemSlots(item).includes(slot));
		return items;
	}

	getEnchants(slot: ItemSlot): Array<Enchant> {
		return this.enchantsBySlot[slot] || [];
	}

	getGems(socketColor?: GemColor): Array<Gem> {
		let gems = Object.values(this.gems);
		if (socketColor) {
			gems = gems.filter(gem => gemEligibleForSocket(gem, socketColor));
		}
		return gems;
	}

	getMatchingGems(socketColor: GemColor): Array<Gem> {
		return Object.values(this.gems).filter(gem => gemMatchesSocket(gem, socketColor));
	}

	lookupItemSpec(itemSpec: ItemSpec): EquippedItem | null {
		const item = this.items[itemSpec.id];
		if (!item)
			return null;

		let enchant: Enchant | null = null;
		if (itemSpec.enchant) {
			const slots = getEligibleItemSlots(item);
			for (let i = 0; i < slots.length; i++) {
				enchant = (this.enchantsBySlot[slots[i]] || [])
						.find(enchant => [enchant.effectId, enchant.itemId, enchant.spellId].includes(itemSpec.enchant)) || null;
				if (enchant) {
					break;
				}
			}
		}

		const gems = itemSpec.gems.map(gemId => this.gems[gemId] || null);

		return new EquippedItem(item, enchant, gems);
	}

	lookupEquipmentSpec(equipSpec: EquipmentSpec): Gear {
		// EquipmentSpec is supposed to be indexed by slot, but here we assume
		// it isn't just in case.
		const gearMap: Partial<Record<ItemSlot, EquippedItem | null>> = {};

		equipSpec.items.forEach(itemSpec => {
			const item = this.lookupItemSpec(itemSpec);
			if (!item)
				return;

			const itemSlots = getEligibleItemSlots(item.item);

			const assignedSlot = itemSlots.find(slot => !gearMap[slot]);
			if (assignedSlot == null)
				throw new Error('No slots left to equip ' + Item.toJsonString(item.item));

			gearMap[assignedSlot] = item;
		});

		return new Gear(gearMap);
	}

	static async getItemIconData(itemId: number): Promise<IconData> {
		const db = await Database.get();
		return db.itemIcons[itemId] || IconData.create();
	}

	static async getSpellIconData(spellId: number): Promise<IconData> {
		const db = await Database.get();
		return db.spellIcons[spellId] || IconData.create();
	}

	//private static async getWowheadTooltipDataHelper(id: number, tooltipPostfix: string, cache: Map<number, Promise<any>>): Promise<any> {
	//	if (!cache.has(id)) {
	//		const url = `https://wowhead.com/wotlk/tooltip/${tooltipPostfix}/${id}`;
	//		try {
	//			const response = await fetch(url);
	//			cache.set(id, response.json());
	//		} catch (e) {
	//			console.error('Error while fetching url: ' + url + '\n\n' + e);
	//			cache.set(id, Promise.resolve({
	//				name: '',
	//				icon: '',
	//				tooltip: '',
	//			}));
	//		}
	//	}

	//	return cache.get(id) as Promise<any>;
	//}
}
