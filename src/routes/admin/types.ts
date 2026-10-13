export type AdminIdentity = {
	provider: string;
	email: string;
	username?: string;
	linkedAt: string;
	lastLogin: string;
};

export type AdminPasskey = {
	id: string;
	name: string;
	createdAt: string;
	lastUsed?: string;
};

export type AdminUser = {
	id: string;
	name: string;
	email: string;
	createdAt: string;
	lastLogin: string;
	profilePicture?: string;
	identities: AdminIdentity[] | null;
	passkeys: AdminPasskey[] | null;
	role?: string;
};

export type AuditEntry = {
	id: string;
	action: string;
	createdAt: string;
	userId?: string;
	ip?: string;
	userAgent?: string;
	requestId?: string;
	target?: { type: string; id: string };
	meta?: Record<string, unknown>;
};

const dateFormat = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' });

export const formatDate = (value: string | undefined) => {
	const date = value ? new Date(value) : null;
	return date && date.getFullYear() > 1 ? dateFormat.format(date) : 'Never';
};
