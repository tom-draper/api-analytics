
export type Candidate = {
    name: string,
    regex: RegExp,
    matches: number
}

export function maintainCandidates(indexUpdated: number, candidates: Candidate[]) {
    // Updated candidate with matches incremented, now need to shift this candidate
    // along to before its original group to maintain sorted order.
    //           v (+1)                          j <-> i            i     j   
    // [6, 5, 5, 5, 1] -> [6, 5, 5, 6, 1] -> [6, 5, 5, 6, 1] -> [6, 6, 5, 5, 1]

    const matches = candidates[indexUpdated].matches;
    // Find index to move updated element to
    let j = indexUpdated;
    while (j > 0 && matches > candidates[j - 1].matches) {
        j--
    }

    // Swap elements to sort candidates
    if (j < indexUpdated) {
        [candidates[indexUpdated], candidates[j]] = [candidates[j], candidates[indexUpdated]]
    }
}