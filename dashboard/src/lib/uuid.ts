export default function formatUUID(userID: string): string {
  return `${userID.slice(0, 8)}-${userID.slice(8, 12)}-${userID.slice(
    12,
    16,
  )}-${userID.slice(16, 20)}-${userID.slice(20)}`
}
