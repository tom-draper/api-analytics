
export type Candidate = {name: string, regex: RegExp, matches: number}

export function maintainCandidates(indexUpdated: number, candidates: Candidate[]) {
    let j = indexUpdated;
    while (j > 0 && count > candidates[j - 1].matches) {
        j--
    }
    if (j < indexUpdated) {
        [candidates[indexUpdated], candidates[j]] = [candidates[j], candidates[indexUpdated]]
    }
}